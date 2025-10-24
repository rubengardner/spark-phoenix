# 🔥 Project Brief: The Phoenix Spark

## 🎯 Core Goal

To build a robust, high-performance Elixir/Phoenix application that operates as a real-time visualization server. The system must accept simultaneous, parameterized HTTP POST requests on coordinate-based endpoints and instantly relay the instruction to a web client using Phoenix Channels, which then triggers a client-side visual animation.

## ✨ Key Features

| ID       | Feature Description                                                                                                                                                                                                                                | Elixir/Phoenix Component Focus |
| :------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :----------------------------- |
| **F-01** | **High-Resolution Coordinate Endpoints:** Create 262,144 unique, addressable API endpoints for a 512x512 grid (e.g., `/api/123/456`) using parameterized routes in the Phoenix Router. | `Router`, `Controller`         |
| **F-02** | **Sparkle Event Parameterization:** The API endpoint must accept a JSON body containing parameters to control the visual effect: `color` (HSL list), `radius` (float), `transparency` (float), and `time_to_grow` (integer, in milliseconds). | `Controller`, `SparkEvent`     |
| **F-03** | **Real-Time Broadcast Engine:** The Controller must publish the spark parameters immediately after validation using Phoenix PubSub to a dedicated Phoenix Channel that the frontend subscribes to.                                                 | `PubSub`, `Channel`            |
| **F-04** | **Frontend Visualizer:** A web-based, full-screen canvas display that scales the 512x512 coordinate space to the viewport. It connects via WebSocket and renders spark animations using the Canvas API and a `requestAnimationFrame` loop. | `JavaScript`, `Phoenix.Channel` |
| **F-05** | **External Client Integration:** The system must be designed to be exclusively "bombarded" by an external Python client using `curl` or the `requests` library, simulating a high-frequency, logic-driven load.                                    | `Deployment`, `API Design`     |

## 🏆 Success Metrics

- **API Responsiveness:** The Phoenix API handler must successfully return a `202 Accepted` status code for all requests within **50 milliseconds** under a simulated load of 10 requests per second.
- **Real-Time Latency:** The total delay, from the successful API response being sent to the Python client to the visible start of the animation on the web display, must be consistently less than **100 milliseconds**.
- **Functional Coverage:** The full 512x512 coordinate space must be addressable via the API, with events correctly triggering animations at their corresponding pixel coordinates on the canvas.
