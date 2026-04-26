<template>
  <section class="analysis">
    <header class="analysis-header">
      <div>
        <h1>Game Analysis</h1>
        <p v-if="game">Game {{ game.ID }} • {{ statusLabel }}</p>
      </div>
      <div v-if="game?.Result" class="result-pill">{{ game.Result }}</div>
    </header>

    <section v-if="loading" class="loading-state">
      <p>Loading game logs…</p>
    </section>
    <section v-else-if="errorMessage" class="loading-state">
      <p>{{ errorMessage }}</p>
    </section>

    <template v-else>
      <div class="summary-grid">
        <article>
          <h2>Winner</h2>
          <p>{{ winnerLabel }}</p>
          <small v-if="game?.WinCondition">Condition: {{ game.WinCondition }}</small>
        </article>
        <article>
          <h2>Total events</h2>
          <p>{{ logs.length }}</p>
        </article>
        <article>
          <h2>Game length</h2>
          <p>{{ durationLabel }}</p>
        </article>
        <article>
          <h2>Turns tracked</h2>
          <p>{{ maxTurnNumber }}</p>
        </article>
      </div>

      <section class="chart-section">
        <header>
          <h2>Life totals</h2>
          <p>Tracked from game logs (life change events).</p>
        </header>
        <div ref="lifeChartEl" class="chart" />
      </section>

      <section class="chart-section">
        <header>
          <h2>Event volume</h2>
          <p>All log events bucketed over time.</p>
        </header>
        <div ref="eventChartEl" class="chart" />
      </section>

      <section class="event-list">
        <header>
          <h2>Key events</h2>
          <p>Latest 20 events in chronological order.</p>
        </header>
        <ul>
          <li v-for="event in recentEvents" :key="event.ID">
            <span class="event-time">{{ formatEventTime(event.EventTime) }}</span>
            <span class="event-type">{{ event.Type }}</span>
            <span class="event-actor" v-if="event.Actor">by {{ event.Actor }}</span>
          </li>
        </ul>
      </section>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { apolloClient } from '../services/apollo';
import { GAME_LOGS_QUERY, GET_GAME_QUERY } from '../graphql/queries';
import uPlot from 'uplot';
import 'uplot/dist/uPlot.min.css';

interface GameLogEvent {
  ID: string;
  GameID: string;
  EventTime: string;
  Type: string;
  Actor?: string | null;
  Payload?: string | null;
}

interface GameSummary {
  ID: string;
  Result?: string | null;
  Status?: string | null;
  WinnerIDs?: string[] | null;
  WinCondition?: string | null;
  Turn?: { Number?: number | null };
  Players: { ID?: string; Username: string; Boardstate?: { Life?: number | null } }[];
}

const route = useRoute();
const game = ref<GameSummary | null>(null);
const logs = ref<GameLogEvent[]>([]);
const loading = ref(true);
const errorMessage = ref<string | null>(null);

const lifeChartEl = ref<HTMLDivElement | null>(null);
const eventChartEl = ref<HTMLDivElement | null>(null);
let lifePlot: uPlot | null = null;
let eventPlot: uPlot | null = null;

const orderedEvents = computed(() => {
  return [...logs.value].sort((a, b) => {
    const taRaw = new Date(a.EventTime).getTime();
    const tbRaw = new Date(b.EventTime).getTime();
    const ta = Number.isFinite(taRaw) ? taRaw : 0;
    const tb = Number.isFinite(tbRaw) ? tbRaw : 0;
    if (ta !== tb) return ta - tb;
    return Number(a.ID) - Number(b.ID);
  });
});

const recentEvents = computed(() => orderedEvents.value.slice(-20));
const statusLabel = computed(() => game.value?.Status ?? 'IN_PROGRESS');
const winnerLabel = computed(() => {
  if (!game.value?.WinnerIDs?.length) return game.value?.Result ?? '—';
  const lookup = new Map((game.value?.Players ?? []).map(p => [p.ID, p.Username]));
  const names = game.value.WinnerIDs.map(id => lookup.get(id) ?? id);
  return names.join(', ');
});

const maxTurnNumber = computed(() => {
  const turns = orderedEvents.value
    .filter(e => e.Type === 'TURN_ADVANCED')
    .map(e => {
      try {
        const payload = e.Payload ? JSON.parse(e.Payload) : null;
        return Number(payload?.to?.number ?? 0);
      } catch {
        return 0;
      }
    });
  const max = Math.max(0, ...(turns.length ? turns : [0]));
  return max || game.value?.Turn?.Number || 0;
});

const durationLabel = computed(() => {
  if (orderedEvents.value.length < 2) return '—';
  const start = new Date(orderedEvents.value[0].EventTime).getTime();
  const end = new Date(orderedEvents.value[orderedEvents.value.length - 1].EventTime).getTime();
  if (!Number.isFinite(start) || !Number.isFinite(end)) return '—';
  const ms = Math.max(0, end - start);
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  return `${minutes}m ${seconds}s`;
});

function formatEventTime(ts: string) {
  const date = new Date(ts);
  if (Number.isNaN(date.getTime())) return '—';
  return date.toLocaleTimeString();
}

function buildLifeSeries(events: GameLogEvent[]) {
  const players = game.value?.Players ?? [];
  const names = players.map(p => p.Username);
  const lifeByPlayer = new Map<string, number>();
  for (const p of players) {
    lifeByPlayer.set(p.Username, p.Boardstate?.Life ?? 40);
  }

  const startEvent = events.find(e => e.Type === 'GAME_CREATED');
  if (startEvent?.Payload) {
    try {
      const payload = JSON.parse(startEvent.Payload);
      const initialPlayers: { username: string; life: number }[] = payload?.players ?? [];
      for (const p of initialPlayers) {
        lifeByPlayer.set(p.username, p.life);
      }
    } catch {}
  }

  const x: number[] = [];
  const series = names.map(() => [] as number[]);
  let startTime = 0;
  if (events.length > 0) {
    const raw = new Date(events[0].EventTime).getTime();
    startTime = Number.isFinite(raw) ? raw : 0;
  }

  function pushPoint(time: number) {
    x.push(time);
    names.forEach((name, idx) => {
      series[idx].push(lifeByPlayer.get(name) ?? 0);
    });
  }

  if (events.length > 0) {
    pushPoint(0);
  }

  let fallbackIndex = 0;
  for (const event of events) {
    if (event.Type !== 'LIFE_CHANGED') continue;
    if (!event.Payload) continue;
    try {
      const payload = JSON.parse(event.Payload);
      const user = payload?.user;
      const to = Number(payload?.to);
      if (user && Number.isFinite(to)) {
        lifeByPlayer.set(user, to);
        const raw = new Date(event.EventTime).getTime();
        const time = Number.isFinite(raw)
          ? Math.max(0, (raw - startTime) / 1000)
          : fallbackIndex++ * 5;
        pushPoint(time);
      }
    } catch {}
  }

  return { x, series, names };
}

function buildEventVolumeSeries(events: GameLogEvent[]) {
  if (!events.length) return { x: [], series: [] as number[] };
  const startRaw = new Date(events[0].EventTime).getTime();
  const startTime = Number.isFinite(startRaw) ? startRaw : 0;
  const bucketSize = 60; // seconds
  const buckets = new Map<number, number>();
  for (const event of events) {
    const raw = new Date(event.EventTime).getTime();
    const offset = Number.isFinite(raw)
      ? Math.max(0, Math.floor((raw - startTime) / 1000))
      : 0;
    const bucket = Math.floor(offset / bucketSize);
    buckets.set(bucket, (buckets.get(bucket) ?? 0) + 1);
  }
  const maxBucket = Math.max(0, ...buckets.keys());
  const x: number[] = [];
  const series: number[] = [];
  for (let i = 0; i <= maxBucket; i++) {
    x.push(i * bucketSize);
    series.push(buckets.get(i) ?? 0);
  }
  return { x, series };
}

function renderLifeChart() {
  if (!lifeChartEl.value) return;
  lifePlot?.destroy();
  const { x, series, names } = buildLifeSeries(orderedEvents.value);
  if (!x.length) return;
  const data: uPlot.AlignedData = [
    Float64Array.from(x),
    ...series.map(points => Float64Array.from(points)),
  ];
  lifePlot = new uPlot({
    width: lifeChartEl.value.clientWidth,
    height: 260,
    series: [
      { label: 'Time (s)' },
      ...names.map((name, idx) => ({
        label: name,
        stroke: ['#8bd5ff', '#ffb347', '#9dffb0', '#ff9de2'][idx % 4],
        width: 2,
      })),
    ],
    axes: [
      { label: 'Seconds' },
      { label: 'Life' },
    ],
  }, data, lifeChartEl.value);
}

function renderEventChart() {
  if (!eventChartEl.value) return;
  eventPlot?.destroy();
  const { x, series } = buildEventVolumeSeries(orderedEvents.value);
  if (!x.length) return;
  const data: uPlot.AlignedData = [
    Float64Array.from(x),
    Float64Array.from(series),
  ];
  eventPlot = new uPlot({
    width: eventChartEl.value.clientWidth,
    height: 200,
    series: [
      { label: 'Time (s)' },
      { label: 'Events', stroke: '#caa4ff', fill: 'rgba(202,164,255,0.2)' },
    ],
    axes: [
      { label: 'Seconds' },
      { label: 'Events / min' },
    ],
  }, data, eventChartEl.value);
}

function rerenderCharts() {
  renderLifeChart();
  renderEventChart();
}

onMounted(async () => {
  loading.value = true;
  errorMessage.value = null;
  try {
    const gameID = route.params.id as string;
    const gameResponse = await apolloClient.query({ query: GET_GAME_QUERY, variables: { gameID }, fetchPolicy: 'network-only' });
    game.value = gameResponse.data?.getGame ?? null;

    const logResponse = await apolloClient.query({ query: GAME_LOGS_QUERY, variables: { gameID, offset: 0, limit: 2000 }, fetchPolicy: 'network-only' });
    logs.value = logResponse.data?.gameLogs ?? [];
  } catch (error) {
    console.error('[analysis] failed to load game logs', error);
    errorMessage.value = 'Unable to load analysis';
  } finally {
    loading.value = false;
  }

  rerenderCharts();
  window.addEventListener('resize', rerenderCharts);
});

onBeforeUnmount(() => {
  lifePlot?.destroy();
  eventPlot?.destroy();
  window.removeEventListener('resize', rerenderCharts);
});
</script>

<style scoped lang="scss">
.analysis {
  display: grid;
  gap: 1.5rem;
}

.analysis-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.result-pill {
  padding: 0.4rem 0.8rem;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background: rgba(255, 255, 255, 0.06);
  font-weight: 600;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 1rem;
}

.summary-grid article {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1rem;
}

.summary-grid h2 {
  font-size: 0.9rem;
  margin-bottom: 0.4rem;
}

.summary-grid p {
  font-size: 1.4rem;
  margin: 0;
}

.chart-section {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 18px;
  padding: 1rem;
}

.chart {
  margin-top: 1rem;
  min-height: 200px;
}

.event-list {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 18px;
  padding: 1rem;
}

.event-list ul {
  list-style: none;
  padding: 0;
  margin: 1rem 0 0;
  display: grid;
  gap: 0.4rem;
}

.event-list li {
  display: flex;
  gap: 0.6rem;
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.8);
}

.event-time {
  color: rgba(255, 255, 255, 0.6);
  min-width: 90px;
}

.event-type {
  font-weight: 600;
}

.event-actor {
  color: rgba(255, 255, 255, 0.7);
}
</style>
