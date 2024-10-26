# k8-mutating-webhook
K8 Mutating webhook project that adds an initContainer to the pod spec to check for Java make and block running Oracle images.

Below are the contents of the project.
1. Mutating webhook yaml (manifest_yamls/webhook.yaml):
   Creates a mutating webhook configuration object on the cluster the YAML is applied.
   This webhook will be invoked by K8 when a pod is created in a namespace that matches the label name 'webhook-poc'
   The pod will fail to be created if any issues in reaching out to the webhook server via service 'pod-mutator-service' as the 'failurePolicy' is set to 'Fail'. This way we are assuring that no pods will be secheduled untill the java make is validated.
   Make sure the CA bundle is updated accordingly to have a proper handshake established between the webhook and the webhook-server. You can create selfsiged certificate for this.
   
2. Service yaml (manifest_yamls/service-webhook-server.yaml):
   Creates a service listening on 443 that routes traffic to the backend webhook-server pod.
   
3. Secrets yaml (manifest_yamls/secret-webhook-cert-key.yaml):
   Creates a secret with couple of entries, one to store the certificate with the CN and SAN as 'pod-mutator-service.webhook-poc.svc' (@@servicename@@-@@namespace@@.svc) and the other to store the private key. Both the cert and key will be mounted to the webhook-server pod to have our HTTP server (webhook-server pod) load them during startup.

4. Webhook server (manifest_yamls/deployment-webhook-server.yaml):
   Creates a deplyment object that launches the webhook-server pod, this is where all the action happens. The pod hosts a HTTP server listens on port 8443 and only handles POST calls on /mutate path/resource (make sure you configure the same path on the mutating webhook yaml).
   K8 addmission controller invokes the webhook and sends a POST call with an admission request as data, which contains the pod information that user is trying to deploy in the namespace.
   The HTTP server creates a patch that adds an initContainer that uses the same image as the main container and checks for java make on the image and gracefully exits (exit 0) when no Oracle java is found and 'exit 1' when Oracle java is found, when the initContainer exits with error (exit 1) initContainer is marked as failed, where the main container does not comeup in the first place. Thus blocking any pod to run that uses Oracle java image.

5. Dockerfile:
   The docker file that creats the image that builds the code webhook-server.go
   It has the commands to configure any priviate repo to pull the dependency packages and any authentication needed against the repo. Make sure you pass the build arguemnt for the token while performing a docker build.

6. webhook-server.go:
   Source code for the HTTP server that handle the POST calls on /mutate resource and sends back a patch to add an initContainer to the pod spec as a response to admission controller.
   
   
   
