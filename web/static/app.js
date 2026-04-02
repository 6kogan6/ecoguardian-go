const serviceStatusEl = document.getElementById("serviceStatus");
const serviceVersionEl = document.getElementById("serviceVersion");
const healthStatusEl = document.getElementById("healthStatus");

const sensorsCountEl = document.getElementById("sensorsCount");
const readingsCountEl = document.getElementById("readingsCount");
const alertsCountEl = document.getElementById("alertsCount");

const sensorsBadgeEl = document.getElementById("sensorsBadge");
const readingsBadgeEl = document.getElementById("readingsBadge");
const alertsBadgeEl = document.getElementById("alertsBadge");

const sensorsListEl = document.getElementById("sensorsList");
const readingsListEl = document.getElementById("readingsList");
const alertsListEl = document.getElementById("alertsList");
const logBoxEl = document.getElementById("logBox");

const sensorForm = document.getElementById("sensorForm");
const readingForm = document.getElementById("readingForm");

const readingSensorIdEl = document.getElementById("readingSensorId");

const refreshBtn = document.getElementById("refreshBtn");
const exportBtn = document.getElementById("exportBtn");

const simulatorBadgeEl = document.getElementById("simulatorBadge");
const simulatorStatusEl = document.getElementById("simulatorStatus");
const simulatorIntervalEl = document.getElementById("simulatorInterval");
const simulatorLastRunEl = document.getElementById("simulatorLastRun");
const simulatorLastBatchEl = document.getElementById("simulatorLastBatch");

const startSimulatorBtn = document.getElementById("startSimulatorBtn");
const stopSimulatorBtn = document.getElementById("stopSimulatorBtn");
const generateOnceBtn = document.getElementById("generateOnceBtn");

function addLog(message) {
  const entry = document.createElement("div");
  entry.className = "log-entry";
  entry.textContent = `${new Date().toLocaleString("ru-RU")} — ${message}`;
  logBoxEl.prepend(entry);

  while (logBoxEl.children.length > 12) {
    logBoxEl.removeChild(logBoxEl.lastChild);
  }
}

async function request(url, options = {}) {
  const response = await fetch(url, options);
  let data = null;

  try {
    data = await response.json();
  } catch (error) {
    data = null;
  }

  if (!response.ok) {
    const errorMessage = data && data.error ? data.error : "Request failed";
    throw new Error(errorMessage);
  }

  return data;
}

function emptyState(text) {
  return `<div class="list-empty">${text}</div>`;
}

function formatDate(value) {
  if (!value) {
    return "—";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString("ru-RU");
}

function levelClass(level, status) {
  if (status === "resolved") {
    return "level-resolved";
  }
  if (level === "danger") {
    return "level-danger";
  }
  return "level-warning";
}

async function loadServiceInfo() {
  const [info, health] = await Promise.all([
    request("/api/info"),
    request("/health"),
  ]);

  serviceStatusEl.textContent = info.status;
  serviceVersionEl.textContent = info.version;
  healthStatusEl.textContent = health.status;
}

async function loadSummary() {
  const summary = await request("/api/dashboard/summary");

  sensorsCountEl.textContent = summary.sensors_count;
  readingsCountEl.textContent = summary.readings_count;
  alertsCountEl.textContent = summary.active_alerts_count;
}

async function loadSimulatorStatus() {
  const status = await request("/api/simulator/status");

  simulatorBadgeEl.textContent = status.running ? "running" : "stopped";
  simulatorStatusEl.textContent = status.running ? "Работает" : "Остановлен";
  simulatorIntervalEl.textContent = `${status.interval_seconds} сек`;
  simulatorLastRunEl.textContent = formatDate(status.last_generated_at);
  simulatorLastBatchEl.textContent = `${status.last_generated_count} показаний / ${status.last_generated_alerts} тревог`;
}

async function loadSensors() {
  const sensors = await request("/api/sensors");

  sensorsBadgeEl.textContent = `${sensors.length} шт.`;
  readingSensorIdEl.innerHTML = "";

  if (sensors.length === 0) {
    sensorsListEl.innerHTML = emptyState(
      "Пока нет датчиков. Сначала создай хотя бы один.",
    );
    const option = document.createElement("option");
    option.value = "";
    option.textContent = "Нет доступных датчиков";
    readingSensorIdEl.appendChild(option);
    return;
  }

  sensorsListEl.innerHTML = sensors
    .map(
      (sensor) => `
    <div class="card-item">
      <div class="card-top">
        <div>
          <h3 class="card-title">${sensor.name}</h3>
          <p class="card-subtitle">Локация: ${sensor.location}</p>
        </div>
        <span class="pill">${sensor.sensor_type}</span>
      </div>
      <div class="muted">Создан: ${formatDate(sensor.created_at)}</div>
    </div>
  `,
    )
    .join("");

  sensors.forEach((sensor) => {
    const option = document.createElement("option");
    option.value = String(sensor.id);
    option.textContent = `${sensor.name} (${sensor.location})`;
    readingSensorIdEl.appendChild(option);
  });
}

async function loadReadings() {
  const readings = await request("/api/readings/latest");

  readingsBadgeEl.textContent = `${readings.length} записей`;

  if (readings.length === 0) {
    readingsListEl.innerHTML = emptyState("Показаний пока нет.");
    return;
  }

  readingsListEl.innerHTML = readings
    .map(
      (reading) => `
    <div class="card-item">
      <div class="card-top">
        <div>
          <h3 class="card-title">${reading.sensor ? reading.sensor.name : "Sensor #" + reading.sensor_id}</h3>
          <p class="card-subtitle">Время: ${formatDate(reading.recorded_at)}</p>
        </div>
      </div>
      <div class="meta-grid">
        <div class="meta-box">
          <div class="meta-label">PM2.5</div>
          <div class="meta-value">${reading.pm25 ?? 0}</div>
        </div>
        <div class="meta-box">
          <div class="meta-label">CO2</div>
          <div class="meta-value">${reading.co2 ?? 0}</div>
        </div>
        <div class="meta-box">
          <div class="meta-label">Шум</div>
          <div class="meta-value">${reading.noise ?? 0}</div>
        </div>
        <div class="meta-box">
          <div class="meta-label">Темп.</div>
          <div class="meta-value">${reading.temperature ?? 0}</div>
        </div>
        <div class="meta-box">
          <div class="meta-label">Влажн.</div>
          <div class="meta-value">${reading.humidity ?? 0}</div>
        </div>
        <div class="meta-box">
          <div class="meta-label">pH</div>
          <div class="meta-value">${reading.water_ph ?? 0}</div>
        </div>
      </div>
    </div>
  `,
    )
    .join("");
}

async function loadAlerts() {
  const alerts = await request("/api/alerts");
  const activeCount = alerts.filter(
    (alert) => alert.status === "active",
  ).length;

  alertsBadgeEl.textContent = `${activeCount} активных`;

  if (alerts.length === 0) {
    alertsListEl.innerHTML = emptyState("Тревог пока нет.");
    return;
  }

  alertsListEl.innerHTML = alerts
    .map(
      (alert) => `
    <div class="card-item">
      <div class="card-top">
        <div>
          <h3 class="card-title">${alert.sensor ? alert.sensor.name : "Sensor #" + alert.sensor_id}</h3>
          <p class="card-subtitle">${alert.message}</p>
        </div>
        <span class="level ${levelClass(alert.level, alert.status)}">${alert.status === "resolved" ? "resolved" : alert.level}</span>
      </div>
      <div class="muted">Создано: ${formatDate(alert.created_at)}</div>
      ${
        alert.status === "active"
          ? `<div style="margin-top: 12px;"><button class="button button-danger button-small" onclick="resolveAlert(${alert.id})">Закрыть тревогу</button></div>`
          : ""
      }
    </div>
  `,
    )
    .join("");
}

async function loadAll() {
  try {
    await Promise.all([
      loadServiceInfo(),
      loadSummary(),
      loadSimulatorStatus(),
      loadSensors(),
      loadReadings(),
      loadAlerts(),
    ]);
  } catch (error) {
    addLog(`Ошибка загрузки данных: ${error.message}`);
  }
}

sensorForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const payload = {
    name: document.getElementById("sensorName").value.trim(),
    location: document.getElementById("sensorLocation").value.trim(),
    sensor_type: document.getElementById("sensorType").value,
  };

  try {
    const created = await request("/api/sensors", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    sensorForm.reset();
    document.getElementById("sensorType").value = "air";

    addLog(`Создан датчик: ${created.name}`);
    await loadAll();
  } catch (error) {
    addLog(`Не удалось создать датчик: ${error.message}`);
  }
});

readingForm.addEventListener("submit", async (event) => {
  event.preventDefault();

  const sensorId = Number(readingSensorIdEl.value);
  if (!sensorId) {
    addLog("Сначала создай датчик.");
    return;
  }

  const payload = {
    sensor_id: sensorId,
    pm25: Number(document.getElementById("pm25").value || 0),
    co2: Number(document.getElementById("co2").value || 0),
    noise: Number(document.getElementById("noise").value || 0),
    temperature: Number(document.getElementById("temperature").value || 0),
    humidity: Number(document.getElementById("humidity").value || 0),
    water_ph: Number(document.getElementById("waterPH").value || 0),
  };

  try {
    const result = await request("/api/readings", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    const sensorName =
      result.reading && result.reading.sensor
        ? result.reading.sensor.name
        : `Sensor #${sensorId}`;
    const alertsCount = Array.isArray(result.created_alerts)
      ? result.created_alerts.length
      : 0;

    addLog(
      `Добавлено показание для ${sensorName}. Создано тревог: ${alertsCount}`,
    );
    await loadAll();
  } catch (error) {
    addLog(`Не удалось отправить показание: ${error.message}`);
  }
});

async function resolveAlert(id) {
  try {
    await request(`/api/alerts/${id}/resolve`, {
      method: "POST",
    });

    addLog(`Тревога #${id} переведена в resolved`);
    await loadAll();
  } catch (error) {
    addLog(`Не удалось закрыть тревогу #${id}: ${error.message}`);
  }
}

window.resolveAlert = resolveAlert;

refreshBtn.addEventListener("click", async () => {
  addLog("Запрошено обновление данных");
  await loadAll();
});

exportBtn.addEventListener("click", async () => {
  try {
    const result = await request("/api/integrations/export", {
      method: "POST",
    });

    addLog(`Экспорт выполнен: ${result.message}`);
  } catch (error) {
    addLog(`Ошибка экспорта: ${error.message}`);
  }
});

startSimulatorBtn.addEventListener("click", async () => {
  try {
    await request("/api/simulator/start", {
      method: "POST",
    });

    addLog("Симулятор запущен");
    await loadAll();
  } catch (error) {
    addLog(`Не удалось запустить симулятор: ${error.message}`);
  }
});

stopSimulatorBtn.addEventListener("click", async () => {
  try {
    await request("/api/simulator/stop", {
      method: "POST",
    });

    addLog("Симулятор остановлен");
    await loadAll();
  } catch (error) {
    addLog(`Не удалось остановить симулятор: ${error.message}`);
  }
});

generateOnceBtn.addEventListener("click", async () => {
  try {
    const result = await request("/api/simulator/generate-once", {
      method: "POST",
    });

    addLog(
      `Сгенерировано: ${result.generated} показаний, тревог: ${result.generated_alerts}`,
    );
    await loadAll();
  } catch (error) {
    addLog(`Ошибка генерации данных: ${error.message}`);
  }
});

let autoRefreshInProgress = false;
setInterval(async () => {
  if (autoRefreshInProgress) {
    return;
  }

  autoRefreshInProgress = true;
  try {
    await loadAll();
  } finally {
    autoRefreshInProgress = false;
  }
}, 5000);

loadAll();
