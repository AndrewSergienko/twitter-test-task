# test-twitter-task

[![codecov](https://codecov.io/github/AndrewSergienko/twitter-test-task/branch/develop/graph/badge.svg?token=EDPw3EC5Mz)](https://codecov.io/github/AndrewSergienko/twitter-test-task)
[![Tests](https://github.com/AndrewSergienko/twitter-test-task/actions/workflows/tests.yml/badge.svg)](https://github.com/AndrewSergienko/twitter-test-task/actions/workflows/tests.yml)

This is a test project named `test-twitter-task` that simulates a simplified Twitter-like application. The project implements the following features:

1. **Get Feed Endpoint**: Fetches existing messages and streams new ones in real-time using HTTP Streaming (Server-Sent Events).
2. **Back Pressure for Message Creation**: Utilizes RabbitMQ to handle back pressure and ensure smooth message creation.
3. **CockroachDB Integration**: A three-node CockroachDB cluster is used as the primary database for storing messages.
4. **Bot for Message Generation**: A bot is implemented to generate messages at a configurable speed to simulate real user activity.

## Getting Started

To set up and run the project, follow these steps:

1. **Create a `.env` file**:
    - Copy the contents of `.env.example` to a new file named `.env`.
    - Update the `.env` file with the appropriate environment variables.

2. **Initialize the Project**:
    - Run the initialization script to set up the environment, generate certificates, and start the necessary services:
      ```bash
      sudo bash scripts/init.sh
      ```

After completing these steps, the project should be up and running, with all services properly configured.
