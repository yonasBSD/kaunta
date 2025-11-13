#!/usr/bin/env -S deno run -A

import * as esbuild from "npm:esbuild@0.20.0";

const tasks = Deno.args.length > 0 ? Deno.args : ["tracker", "vendor"];

async function buildTracker() {
  console.log("Building tracker...");
  await esbuild.build({
    entryPoints: ["tracker/kaunta.js"],
    bundle: true,
    minify: true,
    outfile: "cmd/kaunta/assets/kaunta.min.js",
    target: "es2020",
    format: "iife",
    platform: "browser",
    sourcemap: false,
  });
  console.log("✓ Tracker built");
}

async function buildVendor() {
  console.log("Building vendor...");

  // Ensure output directory exists
  try {
    await Deno.mkdir("cmd/kaunta/assets/dist", { recursive: true });
  } catch (e) {
    // Directory might already exist
  }

  await esbuild.build({
    entryPoints: ["frontend/vendor.ts"],
    bundle: true,
    minify: true,
    outfile: "cmd/kaunta/assets/dist/vendor.js",
    target: "es2020",
    format: "iife",
    platform: "browser",
    sourcemap: false,
    loader: {
      '.png': 'dataurl',
      '.css': 'css',
    },
  });
  console.log("✓ Vendor built");
}

try {
  for (const task of tasks) {
    if (task === "tracker") {
      await buildTracker();
    } else if (task === "vendor") {
      await buildVendor();
    }
  }
  esbuild.stop();
  console.log("✓ Build complete");
} catch (error) {
  console.error("Build failed:", error);
  esbuild.stop();
  Deno.exit(1);
}
