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


## Trigger function

Using the following test, replacing IP with `SERVICE_IP` and Host with `SERVICE_HOST`:
```javascript
function run(msg) {
    request({
        url : 'http://35.232.40.62:80/' + 'run',
        method : 'post',
        parameters : {
            value : msg
        }
    }, function(response) {
        console.log(response);
    }, logger);
}

function request(packet, next, logger) {
    var http = require('request');
    var btoa = require('btoa');

    var options = {
        method: packet.method,
        url : packet.url,
        agentOptions : {
            rejectUnauthorized : false
        },
        headers : {
            'Content-Type' : 'application/json',
            'Host': 'owsk-sample-fn.default.demo-domain.com'
        },
        json : packet.parameters,
    };

    if (packet.auth) {
        options.headers.Authorization = 'Basic ' + btoa(packet.auth);
    }

    http(options, function(error, response, body) {
        if (error) console.log('[error]', error);
        else next(body);
    });
}
function test() {
    run({
        payload : "1 2 3 4 5"
    });
}
test();
```

Run the test:
```console
# node test.js
{ wc: 5 }
```

```