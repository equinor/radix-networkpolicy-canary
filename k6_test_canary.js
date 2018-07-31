import http from "k6/http";
import {
    sleep
} from "k6";

export let options = {
    vus: 300,
    duration: "1800s"
};

export default function () {
    http.get("http://canary.playground-debug-helm-bugs-st4.dev.radix.equinor.com");
    http.get("http://canary.playground-debug-helm-bugs-st4.dev.radix.equinor.com/status");
    http.get("http://canary.playground-debug-helm-bugs-st4.dev.radix.equinor.com/error");
    sleep(1);
};