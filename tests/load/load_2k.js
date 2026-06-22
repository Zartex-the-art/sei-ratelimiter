import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

const allowedCounter = new Counter('requests_allowed');
const blockedCounter = new Counter('requests_blocked');

export const options = {
  scenarios: {
    load: {
      executor: 'constant-arrival-rate',
      rate: 2000,
      timeUnit: '1s',
      duration: '60s',
      preAllocatedVUs: 100,
      maxVUs: 200,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: [
      'p(50)<5',
      'p(95)<10',
      'p(99)<20',
    ],
  },
};

const NODES = [
  'http://localhost:8080',
  'http://localhost:8081',
];

const PAYLOAD = JSON.stringify({
  client_id: 'bench-client',
  algorithm: 'fixed_window',
  limit: 999999,
  window_secs: 3600,
});

export default function () {
  const node = NODES[__VU % 2];

  const res = http.post(`${node}/check`, PAYLOAD, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status 200': (r) => r.status === 200,
    'allowed is bool': (r) => typeof r.json('allowed') === 'boolean',
  });

  if (res.status === 200) {
    const body = res.json();

    if (body.allowed) {
      allowedCounter.add(1);
    } else {
      blockedCounter.add(1);
    }
  }
}

export function handleSummary(data) {
  const dur = data.metrics.http_req_duration.values;

  const p99 = dur['p(99)'] || 0;

  console.log('\n=== 2K RPS BENCHMARK RESULTS ===');
  console.log(`avg: ${(dur.avg || 0).toFixed(2)}ms`);
  console.log(`p90: ${(dur['p(90)'] || 0).toFixed(2)}ms`);
  console.log(`p95: ${(dur['p(95)'] || 0).toFixed(2)}ms`);
  console.log(`p99: ${p99.toFixed(2)}ms`);
  console.log(`max: ${(dur.max || 0).toFixed(2)}ms`);
  console.log(`reqs: ${data.metrics.http_reqs.values.count}`);
  console.log(
    `errs: ${((data.metrics.http_req_failed.values.rate || 0) * 100).toFixed(3)}%`
  );

  if (p99 > 0) {
    if (p99 < 5) {
      console.log('STATUS: EXCELLENT — p99 < 5ms');
    } else if (p99 < 20) {
      console.log('STATUS: ACCEPTABLE — p99 < 20ms');
    } else {
      console.log('STATUS: NEEDS TUNING — p99 >= 20ms');
    }
  }

  return {};
}