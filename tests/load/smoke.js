import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '10s',
  thresholds: {
    http_req_failed:   ['rate<0.01'],
    http_req_duration: ['p(99)<200'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/health');

  check(res, {
    'status is 200':      (r) => r.status === 200,
    'body has status ok': (r) => JSON.parse(r.body).status === 'ok',
  });

  sleep(0.1);
}
