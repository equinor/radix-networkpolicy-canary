/*

k6 run -< k6_test_canary.js

*/

import http from "k6/http";
import {
    sleep
} from "k6";

export let options = {
    vus: 1,
    duration: "1800s"
};

export default function () {

    let appName = "www-radix-canary-golang-dev"
    let clusterName = "dev.dev.radix.equinor.com"

    http.get("https://" + appName + "." + clusterName + "/status");
    http.get("https://" + appName + "." + clusterName + "/error");
    //http.get("https://" + appName + "." + clusterName + "/calculatehashesbcrypt"); // CPU intensive
    http.get("https://" + appName + "." + clusterName + "/calculatehashesscrypt"); // Memory intensive
    sleep(2);
};