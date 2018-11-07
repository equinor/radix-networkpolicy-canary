/*

k6 run -< k6_test_canary.js

*/

import http from "k6/http";
import {
    sleep
} from "k6";

export let options = {
    vus: 10,
    duration: "1800s"
};

export default function () {

    let appName = "www-radix-canary-golang-prod"
    let clusterName = "playground-master-44.dev.radix.equinor.com"

    http.get("https://" + appName + "." + clusterName + "/status");
    http.get("https://" + appName + "." + clusterName + "/error");
    http.get("https://" + appName + "." + clusterName + "/calculatehashesbcrypt");
    http.get("https://" + appName + "." + clusterName + "/calculatehashesscrypt");
    sleep(2);
};