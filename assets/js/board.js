import { Socket } from "phoenix";

let canvas = document.getElementById("board-canvas");
let ctx = canvas.getContext("2d");

const boardSize = 512;

// A list of active sparks to animate
let activeSparks = [];

function initialize() {
  console.log("Canvas initialized.");

  // Connect to the socket
  let socket = new Socket("/socket", { params: {} });
  socket.connect();

  // Join the board channel
  let channel = socket.channel("board:lobby", {});
  channel.join()
    .receive("ok", resp => { console.log("Joined channel successfully", resp) })
    .receive("error", resp => { console.log("Unable to join channel", resp) });

  // Listen for spark events
  channel.on("sparkle_event", payload => {
    drawSpark(payload);
  });
}

function drawSpark(payload) {
    const { x, y, color, radius, transparency, time_to_grow } = payload.data;
    const [h, s, l] = color;

    const startTime = performance.now();

    activeSparks.push({
        x,
        y,
        color: `hsl(${h}, ${s}%, ${l}%)`,
        maxRadius: radius,
        duration: time_to_grow,
        startTime,
        initialOpacity: transparency
    });

    // Start the animation loop if it's not already running
    if (activeSparks.length === 1) {
        requestAnimationFrame(animateSparks);
    }
}

function animateSparks(currentTime) {
    // Clear only the part of the canvas that needs updating if we want to optimize,
    // but clearing the whole canvas is simpler and often fast enough.
    ctx.clearRect(0, 0, boardSize, boardSize);

    let stillActive = [];

    for (const spark of activeSparks) {
        const elapsedTime = currentTime - spark.startTime;
        const progress = Math.min(elapsedTime / spark.duration, 1.0);

        if (progress < 1.0) {
            const currentRadius = spark.maxRadius * progress; // Linear growth
            const currentOpacity = spark.initialOpacity * (1 - progress); // Linear fade

            ctx.fillStyle = spark.color;
            ctx.globalAlpha = currentOpacity;

            ctx.beginPath();
            // Canvas draws pixels centered, so +0.5 aligns with the pixel grid
            ctx.arc(spark.x + 0.5, spark.y + 0.5, currentRadius, 0, 2 * Math.PI);
            ctx.fill();

            stillActive.push(spark);
        }
    }

    activeSparks = stillActive;

    // Keep the loop going if there are sparks to animate
    if (activeSparks.length > 0) {
        requestAnimationFrame(animateSparks);
    } else {
        // Reset alpha after last spark fades
        ctx.globalAlpha = 1.0;
    }
}

// Ensure the DOM is loaded before we start
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initialize);
} else {
    initialize();
}
