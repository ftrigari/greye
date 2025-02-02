import yaml

NUM_SERVICES = 6000
NAMESPACE = "te"
APP_NAME = "printall"
OUTPUT_FILE = "services.yaml"

def generate_service_yaml(index):
    return {
        "apiVersion": "v1",
        "kind": "Service",
        "metadata": {
            "name": f"{APP_NAME}-{index}",
            "namespace": NAMESPACE,
            "annotations": {
                "cm-enabled": "true",
                "cm-intervalSeconds": "30"
            },
            "labels": {
                "app": APP_NAME
            }
        },
        "spec": {
            "ipFamilies": ["IPv4"],
            "ipFamilyPolicy": "SingleStack",
            "ports": [
                {
                    "name": "http",
                    "port": 80,
                    "protocol": "TCP",
                    "targetPort": 5001
                }
            ],
            "selector": {
                "app": APP_NAME
            },
            "sessionAffinity": "None",
            "type": "ClusterIP"
        },
        "status": {
            "loadBalancer": {}
        }
    }

if __name__ == "__main__":
    services = [generate_service_yaml(i) for i in range(1, NUM_SERVICES + 1)]
    with open(OUTPUT_FILE, "w") as f:
        yaml.dump_all(services, f, default_flow_style=False)
    print(f"Generated {NUM_SERVICES} services in {OUTPUT_FILE}")
