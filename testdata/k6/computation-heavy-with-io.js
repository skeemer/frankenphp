import http from 'k6/http';
import { sleep } from 'k6';

const ioLatencyMilliseconds = 5;
const workIterations = 500000;
const outputIterations = 50;

export const options = {
    stages: [
        { duration: '20s', target: 10, },
        { duration: '20s', target: 50 },
        { duration: '20s', target: 0 },
    ],
    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(90)<150'],
    },
};

export default function () {
    http.get(`${__ENV.CADDY_HOSTNAME}/sleep.php?sleep=${ioLatencyMilliseconds}&work=${workIterations}&output=${outputIterations}`);
    //sleep(1);
}