import { Socket } from "phoenix";

let canvas = document.getElementById("board-canvas");
let ctx = canvas.getContext("2d");

const originalWidth = 512;
const originalHeight = 512;

// A list of active sparks to animate
let activeSparks = [];

function resizeCanvas() {
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
}

function initialize() {
  console.log("Canvas initialized.");

  window.addEventListener('resize', resizeCanvas, false);
  resizeCanvas();

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
    handleSparkEvent(payload);
  });
}

function handleSparkEvent(payload) {
    const { x, y, color, radius, transparency, time_to_grow } = payload.data;
    const [h, s, l] = color;

    // Scale the coordinates from the 512x512 system to the current viewport size
    const scaledX = (x / originalWidth) * canvas.width;
    const scaledY = (y / originalHeight) * canvas.height;

    // Scale the radius to be proportional to the viewport size
    const scaleFactor = (canvas.width / originalWidth + canvas.height / originalHeight) / 2;
    const scaledRadius = radius * scaleFactor;

    const startTime = performance.now();

    activeSparks.push({
        x: scaledX,
        y: scaledY,
        color: `hsl(${h}, ${s}%, ${l}%)`,
        maxRadius: scaledRadius,
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
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    let stillActive = [];

    for (const spark of activeSparks) {
        const elapsedTime = currentTime - spark.startTime;
        const progress = Math.max(0, Math.min(elapsedTime / spark.duration, 1.0));

        if (progress < 1.0) {
            const currentRadius = spark.maxRadius * progress; // Linear growth
            const currentOpacity = spark.initialOpacity * (1 - progress); // Linear fade

            ctx.fillStyle = spark.color;
            ctx.globalAlpha = currentOpacity;

            ctx.beginPath();
            ctx.arc(spark.x, spark.y, currentRadius, 0, 2 * Math.PI);
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