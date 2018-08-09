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

    let appName = "canary"
    let clusterName = "playground-v1-1-1-b.dev.radix.equinor.com"

    http.get("http://" + appName + "." + clusterName + "/status");
    http.get("http://" + appName + "." + clusterName + "/error");
    http.get("http://" + appName + "." + clusterName + "/calculatehashesbcrypt");
    http.get("http://" + appName + "." + clusterName + "/calculatehashesscrypt");
    sleep(2);
};