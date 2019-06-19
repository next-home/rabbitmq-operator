package system_tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var _ = Describe("System tests", func() {
	var namespace, podname string
	var clientSet *kubernetes.Clientset

	BeforeEach(func() {
		var err error
		namespace = MustHaveEnv("NAMESPACE")
		podname = MustHaveEnv("PODNAME")
		clientSet, err = createClientSet()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Plugin tests", func() {
		It("can create a test queue and push a message", func() {
			response, err := rabbitmqAlivenessTest(rabbitmqHostName, rabbitmqUsername, rabbitmqPassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Status).To(Equal("ok"))
		})

		It("has required plugins enabled", func() {

			kubectlArgs := []string{
				"-n",
				namespace,
				"exec",
				"-it",
				podname,
				"--",
				"rabbitmq-plugins", "is_enabled",
				"rabbitmq_federation",
				"rabbitmq_federation_management",
				"rabbitmq_management",
				"rabbitmq_peer_discovery_common",
				"rabbitmq_peer_discovery_k8s",
				"rabbitmq_shovel",
				"rabbitmq_shovel_management",
				"rabbitmq_prometheus"}
			kubectlCmd := exec.Command("kubectl", kubectlArgs...)
			err := kubectlCmd.Run()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("ReadinessProbe tests", func() {
		It("stops Rabbitmq app and expects empty endpoints", func() {

			// Run kubectl exec rabbitmqctl stop_app
			err := startStopRabbitmqApp("stop", namespace, podname)
			Expect(err).NotTo(HaveOccurred())

			// Check endpoints and expect they are not ready
			Eventually(func() int {
				return endpointPoller(clientSet, namespace, "rabbitmqcluster-service")
			}, 45, 3).Should(Equal(0))

		})

		AfterEach(func() {
			err := startStopRabbitmqApp("start", namespace, podname)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func startStopRabbitmqApp(startOrStop, namespace, podname string) error {
	kubectlArgs := []string{
		"-n",
		namespace,
		"exec",
		"-it",
		podname,
		"--",
		"rabbitmqctl", fmt.Sprintf("%s_app", startOrStop),
	}

	kubectlCmd := exec.Command("kubectl", kubectlArgs...)
	err := kubectlCmd.Run()
	return err
}

func endpointPoller(clientSet *kubernetes.Clientset, namespace, endpointName string) int {
	endpoints, err := clientSet.CoreV1().Endpoints(namespace).Get(endpointName, v1.GetOptions{})

	if err != nil {
		fmt.Printf("Failed to Get endpoint %s: %v", endpointName, err)
		return -1
	}

	ret := 0
	for _, endpointSubset := range endpoints.Subsets {
		ret = ret + len(endpointSubset.Addresses)
	}

	return ret
}

func rabbitmqAlivenessTest(rabbitmqHostName, rabbitmqUsername, rabbitmqPassword string) (*HealthcheckResponse, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("http://%s:15672/api/aliveness-test/%%2F", rabbitmqHostName)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.SetBasicAuth(rabbitmqUsername, rabbitmqPassword)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to run cluster aliveness test: %+v \n", err)
		return nil, fmt.Errorf("failed aliveness check: %v with api endpoint: %s", err, url)
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read cluster aliveness test: %s \n", err)
		return nil, fmt.Errorf("failed aliveness check: %v with api endpoint: %s", err, url)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Cluster aliveness test failed. Status: %s \n", resp.Status)
		errMessage := fmt.Sprintf("Response code '%d' != '%d'", resp.StatusCode, http.StatusOK)
		return nil, fmt.Errorf("failed aliveness check: %v with api endpoint: %s, error msg: %s", err, url, errMessage)
	}

	healthcheckResponse := &HealthcheckResponse{}
	err = json.Unmarshal(b, healthcheckResponse)
	if err != nil {
		fmt.Printf("Failed to umarshal cluster aliveness test result: %s \n", err)
		return nil, err
	}

	return healthcheckResponse, nil
}

type HealthcheckResponse struct {
	Status string `json:"status"`
}