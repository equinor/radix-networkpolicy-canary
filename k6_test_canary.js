/*

k6 run -< k6_test_canary.js

*/

import http from "k6/http";
import {
    sleep
} from "k6";

export let options = {
    vus: 2,
    duration: "1800s"
};

export default function () {

    const appName = "www-radix-canary-golang-dev"
    const clusterName = "dev.dev.radix.equinor.com"

    const errorProbability = 0.2
    const cpuLoadProbability = 0.2
    const memLoadProbablity = 0.01

    http.get("https://" + appName + "." + clusterName + "/status");
    
    if (Math.random() < errorProbability) {
        http.get("https://" + appName + "." + clusterName + "/error");
    }

    if (Math.random() < cpuLoadProbability) {
        http.get("https://" + appName + "." + clusterName + "/calculatehashesbcrypt"); // CPU intensive
    }
    
    if (Math.random() < memLoadProbablity) {
        http.get("https://" + appName + "." + clusterName + "/calculatehashesscrypt"); // Memory intensive
    }
    
    sleep(2);
};