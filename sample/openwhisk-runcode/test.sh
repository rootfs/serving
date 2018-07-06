export SERVICE_IP=$(kubectl get ing owsk-sample-fn-ingress \
  -o jsonpath="{.status.loadBalancer.ingress[0]['ip']}")
  
export SERVICE_HOST=$(kubectl get ing owsk-sample-fn-ingress \
  -o jsonpath="{.spec.rules[0]['host']}")
echo ${SERVICE_IP} ${SERVICE_HOST}
node test.js
