#!/bin/bash

echo "Testing Code Execution API..."
echo ""

echo "1. Submitting Python code..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d '{
    "language": "python",
    "code": "n = int(input())\nprint(n * 2)",
    "test_cases": [
      {"input": "5", "expected": "10"},
      {"input": "10", "expected": "20"}
    ],
    "time_limit": 5,
    "memory_limit": 256
  }')

echo "Response: $RESPONSE"
SUBMISSION_ID=$(echo $RESPONSE | grep -o '"submission_id":"[^"]*"' | cut -d'"' -f4)
echo "Submission ID: $SUBMISSION_ID"
echo ""

echo "2. Waiting 2 seconds for processing..."
sleep 2
echo ""

echo "3. Checking status..."
curl -s http://localhost:8080/api/status/$SUBMISSION_ID | jq .
echo ""

echo "4. Testing C++ code..."
CPP_RESPONSE=$(curl -s -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d '{
    "language": "cpp",
    "code": "#include <iostream>\nusing namespace std;\nint main() { int n; cin >> n; cout << n * 3 << endl; return 0; }",
    "test_cases": [
      {"input": "4", "expected": "12"}
    ],
    "time_limit": 10,
    "memory_limit": 512
  }')

echo "Response: $CPP_RESPONSE"
CPP_ID=$(echo $CPP_RESPONSE | grep -o '"submission_id":"[^"]*"' | cut -d'"' -f4)
echo "Submission ID: $CPP_ID"
echo ""

echo "5. Waiting 3 seconds for compilation and execution..."
sleep 3
echo ""

echo "6. Checking C++ status..."
curl -s http://localhost:8080/api/status/$CPP_ID | jq .
echo ""

echo "7. Testing health endpoint..."
curl -s http://localhost:8080/health
echo ""
echo ""

echo "Test completed!"
