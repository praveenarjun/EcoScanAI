const form = document.getElementById("scanForm");
const input = document.getElementById("imageInput");
const dropzone = document.getElementById("dropzone");
const dropText = document.getElementById("dropText");
const previewWrap = document.getElementById("previewWrap");
const preview = document.getElementById("preview");
const scanBtn = document.getElementById("scanBtn");

const statusText = document.getElementById("statusText");
const resultContent = document.getElementById("resultContent");
const itemValue = document.getElementById("itemValue");
const materialValue = document.getElementById("materialValue");
const disposalValue = document.getElementById("disposalValue");
const tipsValue = document.getElementById("tipsValue");
const carbonValue = document.getElementById("carbonValue");
const progressBar = document.getElementById("progressBar");
const sourceValue = document.getElementById("sourceValue");
const latencyValue = document.getElementById("latencyValue");
const impactStory = document.getElementById("impactStory");
const storyTitle = document.getElementById("storyTitle");
const storyBody = document.getElementById("storyBody");
const impactScore = document.getElementById("impactScore");
const nextAction = document.getElementById("nextAction");

function setPreview(file) {
  const reader = new FileReader();
  reader.onload = (event) => {
    preview.src = event.target.result;
    previewWrap.hidden = false;
    dropText.textContent = file.name;
  };
  reader.readAsDataURL(file);
}

function parseCarbonToGrams(text) {
  if (!text) {
    return 0;
  }

  const value = String(text).toLowerCase().replace(/,/g, "").trim();
  const match = value.match(/([0-9]+(?:\.[0-9]+)?)\s*(kg|g|grams?)?/);
  if (!match) {
    return 0;
  }

  const amount = Number.parseFloat(match[1]);
  const unit = match[2] || "g";

  if (!Number.isFinite(amount)) {
    return 0;
  }

  if (unit.startsWith("kg")) {
    return amount * 1000;
  }

  return amount;
}

function mapGramsToProgress(grams) {
  const clamped = Math.max(0, Math.min(1000, grams));
  return Math.round((clamped / 1000) * 100);
}

function createImpactNarrative(result, grams, progress) {
  const disposal = String(result.disposal || "").toLowerCase();
  const item = result.item || "this item";

  if (disposal.includes("recycle") || disposal.includes("reuse") || disposal.includes("compost")) {
    const score = Math.min(98, Math.max(65, progress + 35));
    return {
      title: "Planet-positive move unlocked",
      body: `Great call. ${item} can stay in the circular economy instead of ending up as waste.` ,
      score,
      action: "Next action: Share this tip with 1 friend and multiply the impact.",
    };
  }

  if (disposal.includes("hazardous")) {
    const score = Math.min(80, Math.max(45, progress + 18));
    return {
      title: "Safety-first climate action",
      body: `Smart scan. Hazardous items need correct collection points to prevent air, soil, and water pollution.` ,
      score,
      action: "Next action: Locate the nearest hazardous-waste drop-off center.",
    };
  }

  const fallbackScore = Math.min(72, Math.max(38, progress + 22));
  return {
    title: "Better disposal starts with visibility",
    body: `You just prevented guesswork. Knowing how to handle ${item} is the first step toward cleaner neighborhoods.` ,
    score: fallbackScore,
    action: "Next action: Scan another everyday item to build a weekly impact streak.",
  };
}

function renderResult(payload) {
  const result = payload.result || {};
  const grams = parseCarbonToGrams(result.carbon_save);
  const progress = mapGramsToProgress(grams);
  const narrative = createImpactNarrative(result, grams, progress);

  itemValue.textContent = result.item || "Unknown";
  materialValue.textContent = result.material || "Unknown";
  disposalValue.textContent = result.disposal || "Unknown";
  tipsValue.textContent = result.tips || "No tip returned";
  carbonValue.textContent = result.carbon_save || "0g CO2";
  progressBar.style.width = `${progress}%`;
  sourceValue.textContent = `Source: ${payload.source || "ai"}`;
  latencyValue.textContent = `Latency: ${payload.latency_ms || 0} ms`;

  storyTitle.textContent = narrative.title;
  storyBody.textContent = narrative.body;
  impactScore.textContent = `Impact Score: ${narrative.score}/100`;
  nextAction.textContent = narrative.action;
  impactStory.hidden = false;

  statusText.textContent = "Analysis completed. This is challenge-ready insight for real-world behavior change.";
  resultContent.hidden = false;
}

async function submitScan(file) {
  const fd = new FormData();
  fd.append("image", file);

  const response = await fetch("/api/scan", {
    method: "POST",
    body: fd,
  });

  const data = await response.json();

  if (!response.ok) {
    throw new Error(data.error || "Scan failed");
  }

  return data;
}

input.addEventListener("change", () => {
  const file = input.files && input.files[0];
  if (file) {
    setPreview(file);
  }
});

dropzone.addEventListener("dragover", (e) => {
  e.preventDefault();
  dropzone.classList.add("dragover");
});

dropzone.addEventListener("dragleave", () => {
  dropzone.classList.remove("dragover");
});

dropzone.addEventListener("drop", (e) => {
  e.preventDefault();
  dropzone.classList.remove("dragover");
  const file = e.dataTransfer.files && e.dataTransfer.files[0];
  if (file) {
    const transfer = new DataTransfer();
    transfer.items.add(file);
    input.files = transfer.files;
    setPreview(file);
  }
});

form.addEventListener("submit", async (e) => {
  e.preventDefault();

  const file = input.files && input.files[0];
  if (!file) {
    statusText.textContent = "Please choose an image first.";
    return;
  }

  scanBtn.disabled = true;
  scanBtn.textContent = "Analyzing...";
  statusText.textContent = "Analyzing item and generating your Earth Day impact story...";
  resultContent.hidden = true;
  impactStory.hidden = true;

  try {
    const payload = await submitScan(file);
    renderResult(payload);
  } catch (error) {
    statusText.textContent = error.message;
  } finally {
    scanBtn.disabled = false;
    scanBtn.textContent = "Analyze Item";
  }
});
