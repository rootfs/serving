# OpenWhisk Function Example

## Start

```bash
# kubectl create -f sample.yaml
```

## Get endpoint

```bash
export SERVICE_IP=$(kubectl get ing owsk-sample-fn-ingress -o jsonpath="{.status.loadBalancer.ingress[0]['ip']}")
  
export SERVICE_HOST=$(kubectl get ing owsk-sample-fn-ingress -o jsonpath="{.spec.rules[0]['host']}")
```


## Serverless function

As illustrated at the sample, the serverless function used [here](https://github.com/rootfs-dev/serverless-ex) is a nodejs function that exporting `main` function, as the following

```javascript

function foo(msg)
{
    var wc = msg.payload.split(" ").length;
    console.log("wc:", wc);
    return { wc: wc};
}
function main(msg)
{
    return foo(msg);
}

module.exports.main = main;
```

## Test

Run the test:
```console
# sh test.sh
{ wc: 5 }
```

```