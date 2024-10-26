package main

import (
    "encoding/json"
    "encoding/base64"
    "fmt"
    "log"

    "github.com/gin-gonic/gin"
    //metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    admissionv1 "k8s.io/api/admission/v1"
    corev1 "k8s.io/api/core/v1"
)

type AdmissionResponse struct {
    Status string `json:"status"`
}

type AdmissionReview struct {
    Kind       string               `json:"kind"`
    APIVersion string               `json:"apiVersion"`
    Response  AdmissionReviewResponse `json:"response"`
}

type AdmissionReviewResponse struct {
    UID        string          `json:"uid"`
    Allowed    bool            `json:"allowed"`
    Patch      string  `json:"patch"`
    PatchType  string          `json:"patchType"`
    Result     AdmissionResponse `json:"result"`
}


func serveMutatePod(c *gin.Context) {
    // Extract the AdmissionReview object from the request
    var admissionReview admissionv1.AdmissionReview
    if err := c.ShouldBindJSON(&admissionReview); err != nil {
        c.JSON(400, gin.H{"error": "Failed to decode admission review", "details": err.Error()})
        return
    }

    // Log the decoded AdmissionReview (for debugging purposes)
    admissionReviewJSON, err := json.Marshal(admissionReview)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to marshal admission review", "details": err.Error()})
        return
    }
    
    admissionReviewReq := admissionReview.Request
    fmt.Printf("AdmissionReview: %v\n", admissionReviewReq)
    fmt.Printf("")
    fmt.Printf("Decoded AdmissionReview: %s\n", string(admissionReviewJSON))

    // Extract UID from the request
    uid := string(admissionReview.Request.UID)

    // Unmarshal the Pod object from the AdmissionReview
    pod := corev1.Pod{}
    if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
        c.JSON(500, gin.H{"error": "Failed to unmarshal Pod", "details": err.Error()})
        return
    }

    // Log the updated Pod (for debugging)
    updatedPodJSON, err := json.Marshal(pod)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to marshal updated Pod", "details": err.Error()})
        return
    }

    fmt.Printf("Updated Pod with Init Container: %s\n", string(updatedPodJSON))

    // Create a patch operation to add the init container
    if len(pod.Spec.InitContainers) == 0 {
    	patch := []map[string]interface{}{
    		"op":    "add",
    		"path":  "/spec/initContainers",
    		"value": []map[string]interface{}{
    			{
    				"command": []string{"sh", "-c", "echo 'Checking Java version..' ; if java -version 2>&1 | grep -qi oracle; then echo 'Oracle java found, this will be a non-zero exit'; exit 1; else echo 'Oracle java NOT found, exiting graccefully..' ; exit 0; fi",},
    				"image":   pod.Spec.Containers[0].Image,
    				"name":    "java-version-check",
    			},
    		},
    	}
    } else {
    	patch := []map[string]interface{}{
    		"op":    "add",
    		"path":  "/spec/initContainers",
    		"value": []map[string]interface{}{
    			{
    				"command": []string{"sh", "-c", "echo 'Checking Java version..' ; if java -version 2>&1 | grep -qi oracle; then echo 'Oracle java found, this will be a non-zero exit'; exit 1; else echo 'Oracle java NOT found, exiting graccefully..' ; exit 0; fi",},
    				"image":   pod.Spec.Containers[0].Image,
    				"name":    "java-version-check",
    			},
    		},
    	}
    }


    // Marshal the patch into bytes
    patchJSON, err := json.Marshal(patch)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to marshal patch", "details": err.Error()})
        return
    }

    base64EncodedPatch := base64.StdEncoding.EncodeToString(patchJSON)


//fmt.Printf("This is patch: %s\n", patch)


// Create the admission response
admissionResponse := AdmissionReview{
    Kind:       "AdmissionReview",
    APIVersion: "admission.k8s.io/v1",
    Response: AdmissionReviewResponse{
        UID:     uid,
        Allowed: true,
        Patch:   base64EncodedPatch,
        PatchType: "JSONPatch",
        Result: AdmissionResponse{
            Status: "Success",
        },
    },
}


    // Marshal the AdmissionResponse
    admissionResponseBytes, err := json.Marshal(admissionResponse)
    if err != nil {
        c.JSON(500, gin.H{"error": "Error marshalling AdmissionResponse", "details": err.Error()})
        return
    }

    // Send the AdmissionResponse back to Kubernetes
    c.Header("Content-Type", "application/json")
    c.JSON(200, admissionResponse)
    
    fmt.Printf("Admission response sent: %s\n", admissionResponse)
    fmt.Printf("Admission response sent: %s\n", string(admissionResponseBytes))

jsonResponse, _ := json.MarshalIndent(admissionResponse, "", "  ")
fmt.Printf("AdmissionResponse: %s\n", jsonResponse)

}

func main() {
    r := gin.Default()

    // Define a route that listens for POST requests on /mutate
    r.POST("/mutate", serveMutatePod)

    // Run the Gin server with TLS
    err := r.RunTLS(":8443", "/etc/webhook/certs/tls.crt", "/etc/webhook/certs/tls.key")
    if err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}
