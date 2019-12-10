package resource_test

import (
	b64 "encoding/base64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rabbitmqv1beta1 "github.com/pivotal/rabbitmq-for-kubernetes/api/v1beta1"
	"github.com/pivotal/rabbitmq-for-kubernetes/internal/resource"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("AdminSecret", func() {
	var (
		secret             *corev1.Secret
		instance           rabbitmqv1beta1.RabbitmqCluster
		rabbitmqCluster    *resource.RabbitmqResourceBuilder
		adminSecretBuilder *resource.AdminSecretBuilder
	)

	BeforeEach(func() {
		instance = rabbitmqv1beta1.RabbitmqCluster{
			ObjectMeta: v1.ObjectMeta{
				Name:      "a name",
				Namespace: "a namespace",
			},
		}
		rabbitmqCluster = &resource.RabbitmqResourceBuilder{
			Instance: &instance,
		}
		adminSecretBuilder = rabbitmqCluster.AdminSecret()
	})

	Context("Build", func() {
		BeforeEach(func() {
			obj, err := adminSecretBuilder.Build()
			Expect(err).NotTo(HaveOccurred())
			secret = obj.(*corev1.Secret)
		})

		It("creates the secret with correct name and namespace", func() {
			Expect(secret.Name).To(Equal(instance.ChildResourceName("admin")))
			Expect(secret.Namespace).To(Equal("a namespace"))
		})

		It("creates a 'opaque' secret ", func() {
			Expect(secret.Type).To(Equal(corev1.SecretTypeOpaque))
		})

		It("creates a rabbitmq username that is base64 encoded and 24 characters in length", func() {
			username, ok := secret.Data["rabbitmq-username"]
			Expect(ok).NotTo(BeFalse())
			decodedUsername, err := b64.URLEncoding.DecodeString(string(username))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(decodedUsername)).To(Equal(24))

		})

		It("creates a rabbitmq password that is base64 encoded and 24 characters in length", func() {
			password, ok := secret.Data["rabbitmq-password"]
			Expect(ok).NotTo(BeFalse())
			decodedPassword, err := b64.URLEncoding.DecodeString(string(password))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(decodedPassword)).To(Equal(24))
		})

		It("only creates the required labels", func() {
			labels := secret.Labels
			Expect(len(labels)).To(Equal(3))
			Expect(labels["app.kubernetes.io/name"]).To(Equal(instance.Name))
			Expect(labels["app.kubernetes.io/component"]).To(Equal("rabbitmq"))
			Expect(labels["app.kubernetes.io/part-of"]).To(Equal("pivotal-rabbitmq"))
		})
	})

	Context("Build with labels on CR", func() {
		BeforeEach(func() {
			instance.Labels = map[string]string{
				"app.kubernetes.io/foo": "bar",
				"foo":                   "bar",
				"rabbitmq":              "is-great",
				"foo/app.kubernetes.io": "edgecase",
			}

			obj, err := adminSecretBuilder.Build()
			Expect(err).NotTo(HaveOccurred())
			secret = obj.(*corev1.Secret)
		})

		It("has the labels from the CRD on the admin secret", func() {
			testLabels(secret.Labels)
		})

		It("also has the required labels", func() {
			labels := secret.Labels
			Expect(labels["app.kubernetes.io/name"]).To(Equal(instance.Name))
			Expect(labels["app.kubernetes.io/component"]).To(Equal("rabbitmq"))
			Expect(labels["app.kubernetes.io/part-of"]).To(Equal("pivotal-rabbitmq"))
		})
	})

	Context("Update", func() {
		BeforeEach(func() {
			instance = rabbitmqv1beta1.RabbitmqCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "rabbit-labelled",
				},
			}
			instance.Labels = map[string]string{
				"app.kubernetes.io/foo": "bar",
				"foo":                   "bar",
				"rabbitmq":              "is-great",
				"foo/app.kubernetes.io": "edgecase",
			}

			secret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":      instance.Name,
						"app.kubernetes.io/part-of":   "pivotal-rabbitmq",
						"this-was-the-previous-label": "should-be-deleted",
					},
				},
			}
			err := adminSecretBuilder.Update(secret)
			Expect(err).NotTo(HaveOccurred())
		})

		It("adds labels from the CR", func() {
			testLabels(secret.Labels)
		})

		It("restores the default labels", func() {
			labels := secret.Labels
			Expect(labels["app.kubernetes.io/name"]).To(Equal(instance.Name))
			Expect(labels["app.kubernetes.io/component"]).To(Equal("rabbitmq"))
			Expect(labels["app.kubernetes.io/part-of"]).To(Equal("pivotal-rabbitmq"))
		})

		It("deletes the labels that are removed from the CR", func() {
			Expect(secret.Labels).NotTo(HaveKey("this-was-the-previous-label"))
		})
	})
})
