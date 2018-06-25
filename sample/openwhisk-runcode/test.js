/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/*
*  based on https://github.com/apache/incubator-openwhisk-runtime-nodejs/blob/master/core/nodejsActionBase/test.js
*/
var logger = {
    info : console.log,
    error : console.log
}

var getenv = require('getenv');
var host = getenv('SERVICE_HOST');
var ip = getenv('SERVICE_IP');

function run(msg) {
    request({
        url : 'http://' + ip + ':80/' + 'run',
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
            'Host': host
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
