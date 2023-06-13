// Modified from https://k6.io/docs/test-types/load-testing/ and https://k6.io/docs/testing-guides/api-load-testing/
import http from 'k6/http';
import {sleep} from 'k6';

// Test for this number of users
const numUsers = 100;

// HTTP error rate should be less than 1% and 99% of requests should be processed within 0.1 second
export const options = {
    thresholds: {
        http_req_failed: ['rate<0.01'], 
        http_req_duration: ['p(99)<100'], 
    },
    stages: [
        { duration: '30s', target: numUsers },
        { duration: '4m', target: numUsers },
        { duration: '30s', target: 0 },
    ],
};

// Testing routine:
// Step 1: Send hi from a random user to another random user (total of numUsers users)
// Step 2: Retrieve up to 10 latest messages between the two chosen users
export default () => {
    const user1 = Math.floor(Math.random() * numUsers)
    const user2 = Math.floor(Math.random() * numUsers)
    http.post('http://localhost:8080/api/send?sender=' + user1 + '&receiver=' + user2 + '&text=Hi!%20How%20are%20you?');
    http.get('http://localhost:8080/api/pull?chat=' + user1 + ':' + user2);
};
