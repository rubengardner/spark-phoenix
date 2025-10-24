# 🔥 The Phoenix Spark

A real-time, high-performance visualization engine built with Elixir and Phoenix, designed to handle and display high-volume API activity on a dynamic canvas.

---

## 🎬 Demo

![App Demo](assets/demo.gif)

---

This project demonstrates the power of the Elixir, the Phoenix Framework, and Phoenix Channels for building robust, concurrent, and low-latency real-time applications. It functions as a server that ingests a high volume of parameterized API requests and broadcasts them to a frontend visualizer, which renders the events as "sparks" on a full-screen HTML Canvas.

## ✨ Key Features

- **High-Resolution Coordinate Endpoints:** 262,144 unique API endpoints on a 512x512 grid (`/api/:x/:y`) are handled by a single, parameterized Phoenix Router.
- **Rich Event Parameterization:** API requests accept a JSON body to control the visual effect of each spark, including `color` (HSL), `radius`, `transparency`, and animation `duration`.
- **Asynchronous Broadcast Engine:** The controller provides an immediate `202 Accepted` response and delegates the broadcasting to a non-blocking background task for maximum throughput.
- **High-Performance Frontend Visualizer:** A full-screen, borderless HTML Canvas display that dynamically scales the 512x512 coordinate space to the viewport. Animations are rendered using a `requestAnimationFrame` loop for smoothness.
- **Robust Data Validation:** A dedicated `SparkEvent` data contract module uses Ecto changesets to validate all incoming API data for type, presence, and valid ranges.

## 🛠️ Tech Stack

- **Backend:**

  - [Elixir](https://elixir-lang.org/)
  - [Phoenix Framework](https://www.phoenixframework.org/)
  - [Phoenix PubSub](https://hexdocs.pm/phoenix_pubsub/Phoenix.PubSub.html) for broadcasting
  - [Ecto](https://hexdocs.pm/ecto/Ecto.html) Changesets for data validation

- **Frontend:**

  - HTML5 Canvas API
  - JavaScript (ES6+)
  - [Phoenix Channels](https://hexdocs.pm/phoenix/channels.html) for WebSocket communication

- **Tooling:**
  - `esbuild` for JavaScript bundling

## 🚀 Getting Started

### Prerequisites

- Elixir `~> 1.15`
- Erlang/OTP `~> 26`
- Python 3 (for the bombarder script)

### 1. Setup the Backend

First, install the Mix dependencies:

```sh
mix deps.get
```

If this is the first time you are running the project, you may also need to create and migrate your database (although it is not actively used by the core application):

```sh
mix ecto.create
mix ecto.migrate
```

### 2. Run the Server

Start the Phoenix server:

```sh
mix phx.server
```

Now you can visit [`http://localhost:4000`](http://localhost:4000) from your browser to see the canvas.

### 3. Run the Bombarder Script

To see the visualization in action, run the updated Python bombarder script. This script is designed for maximum throughput and visual flair, sending thousands of requests per second with dynamically changing colors, sizes, and animations.

**Example:** Run the bombardment for 30 seconds.

```sh
./scripts/bombard.py --duration 30
```

See all available options by running:

```sh
./scripts/bombard.py --help
```

## 📚 Documentation

For a deeper dive into the project's architecture and requirements, please see the detailed documentation:

- **[Project Brief](./docs/project-brief.md):** High-level goals, features, and success metrics.
- **[Backend Specification](./docs/back-end-spec.md):** A detailed look at the server-side architecture, request flow, and data validation.
- **[UI/UX Specification](./docs/front-end-spec.md):** A detailed look at the frontend canvas, rendering logic, and animation design.
