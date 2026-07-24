<template>
  <!-- Root element is an li to drop into existing lists -->
  <li
    class="card-tile"
    :class="[{ dragging, tapped: isTapped, facedown: isFaceDown, selected, highlight, [`size-${size}`]: true }]"
    :draggable="draggable"
    role="img"
    :aria-label="ariaLabel"
  :aria-grabbed="draggable ? !!dragging : undefined"
    @dragstart="emit('dragstart', $event)"
    @dragend="emit('dragend', $event)"
    @click="emit('click', $event)"
    @dblclick="emit('dblclick', $event)"
    @contextmenu.prevent="emit('contextmenu', $event)"
  >
    <div class="media" :class="displayMode">
      <!-- image/front/back -->
      <template v-if="displayMode !== 'compact'">
        <img
          v-if="!isFaceDown"
          class="image"
          :src="imageSrc"
          :alt="name"
          @error="onImgError"
        />
        <img
          v-else-if="backImageSrc"
          class="image back"
          :src="backImageSrc"
          alt="Card back"
          @error="onImgError"
        />
        <div v-else class="image placeholder" aria-hidden="true">
          <span>Card Back</span>
        </div>
      </template>

      <!-- overlays -->
      <div class="overlays">
        <!-- counters (top-left) -->
        <div v-if="counters?.length" class="counters">
          <span v-for="(c, i) in counters" :key="i" class="counter" :style="{ background: c.color || 'rgba(0,0,0,0.65)' }" :title="c.type">
            <strong>{{ c.count }}</strong>
            <small v-if="c.type">{{ c.type }}</small>
          </span>
        </div>

        <!-- labels (bottom-right) -->
        <div v-if="labels?.length" class="labels">
          <span v-for="(l, i) in normalizedLabels" :key="i" class="label-chip" :style="{ background: l.color || 'rgba(0,0,0,0.55)' }">{{ l.text }}</span>
        </div>
      </div>
    </div>

    <!-- caption/name -->
    <div v-if="showName && displayMode !== 'thumb'" class="label">{{ name }}</div>
  </li>
</template>

<script setup lang="ts">
import { computed } from 'vue';

type Counter = { type?: string; count: number; color?: string };
type Label = { text: string; color?: string } | string;

const props = withDefaults(defineProps<{
  id: string;
  name: string;
  imageSrc?: string;
  backImageSrc?: string;
  faceDown?: boolean;
  tapped?: boolean;
  counters?: Counter[];
  labels?: Label[];
  draggable?: boolean;
  dragging?: boolean;
  displayMode?: 'image' | 'compact' | 'thumb';
  size?: 'xs' | 'sm' | 'md' | 'lg';
  showName?: boolean;
  selected?: boolean;
  highlight?: boolean;
}>(), {
  faceDown: false,
  tapped: false,
  counters: () => [],
  labels: () => [],
  draggable: false,
  dragging: false,
  displayMode: 'image',
  size: 'md',
  showName: true,
  selected: false,
  highlight: false,
});

const emit = defineEmits<{
  (e: 'dragstart', ev: DragEvent): void;
  (e: 'dragend', ev: DragEvent): void;
  (e: 'click', ev: MouseEvent): void;
  (e: 'dblclick', ev: MouseEvent): void;
  (e: 'contextmenu', ev: MouseEvent): void;
}>();

const isFaceDown = computed(() => !!props.faceDown);
const isTapped = computed(() => !!props.tapped);
const ariaLabel = computed(() => {
  const parts = [props.name];
  if (isFaceDown.value) parts.push('(face down)');
  if (isTapped.value) parts.push('(tapped)');
  return parts.join(' ');
});

const normalizedLabels = computed(() =>
  (props.labels || []).map(l => typeof l === 'string' ? { text: l } : l)
);

function onImgError(ev: Event) {
  const el = ev.target as HTMLImageElement;
  el.style.visibility = 'hidden';
}
</script>

<style scoped lang="scss">
/* Sizing via modifier class */
.card-tile {
  display: grid;
  gap: 0.35rem;
  background: linear-gradient(180deg, rgba(255, 244, 237, 0.08), rgba(26, 22, 21, 0.34));
  border: 1px solid rgba(255, 244, 237, 0.11);
  border-radius: 14px;
  padding: 0.4rem;
  cursor: grab;
  position: relative;
  overflow: hidden;
  box-shadow: 0 10px 24px rgba(21, 12, 9, 0.18);
  transition: transform 140ms ease, box-shadow 160ms ease, border-color 140ms ease, background 160ms ease;
}
.card-tile:active { cursor: grabbing; }

.card-tile .label {
  font-size: 0.8rem;
  opacity: 0.95;
  line-height: 1.25;
  font-weight: 500;
  color: var(--vedh-text);
  text-wrap: balance;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  min-height: 2em;
}

.media {
  position: relative;
}

.media::after {
  content: '';
  position: absolute;
  inset: auto 0 0 0;
  height: 34%;
  background: linear-gradient(180deg, rgba(0,0,0,0), rgba(18, 12, 10, 0.38));
  pointer-events: none;
}

.media.image .image,
.media.thumb .image {
  width: 100%;
  aspect-ratio: 0.714;
  object-fit: cover;
  border-radius: 10px;
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.28);
  transition: transform 180ms ease, filter 180ms ease;
}
.media .image.placeholder {
  background: linear-gradient(135deg, rgba(40,40,40,0.9), rgba(22,22,22,0.9));
  color: rgba(255,255,255,0.7);
  display: grid;
  place-items: center;
  min-height: 100%;
}

/* tapped rotates the card image only */
.tapped .media .image {
  transform: rotate(90deg) scale(0.92);
  transform-origin: center;
}

.tapped::before,
.facedown::before {
  content: attr(data-state);
  position: absolute;
  top: 0.7rem;
  right: 0.7rem;
  z-index: 3;
  padding: 0.2rem 0.42rem;
  border-radius: 999px;
  font-size: 0.64rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--vedh-text);
  background: rgba(17, 13, 12, 0.72);
  border: 1px solid rgba(255, 244, 237, 0.16);
}

.tapped { --card-state-label: 'Tapped'; }
.facedown { --card-state-label: 'Face down'; }

.tapped::before { content: 'Tapped'; }
.facedown::before { content: 'Face down'; }

.overlays {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.counters {
  position: absolute;
  top: 6px;
  left: 6px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.counter {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: #fff;
  padding: 2px 6px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 1;
  box-shadow: 0 1px 3px rgba(0,0,0,0.4);
}
.counter strong { font-weight: 700; }
.counter small { opacity: 0.8; }

.labels {
  position: absolute;
  right: 6px;
  bottom: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
.label-chip {
  color: #fff;
  padding: 2px 6px;
  border-radius: 6px;
  font-size: 11px;
  line-height: 1;
  box-shadow: 0 1px 3px rgba(0,0,0,0.35);
}

/* Dragging visuals to match BoardView */
@keyframes pulse-glow {
  0% { box-shadow: 0 6px 18px rgba(133,215,255,0.0); }
  50% { box-shadow: 0 10px 30px rgba(133,215,255,0.25); }
  100% { box-shadow: 0 6px 18px rgba(133,215,255,0.0); }
}
.card-tile.dragging {
  transform: translateY(-6px) scale(1.02);
  animation: pulse-glow 1.2s ease-in-out infinite;
  z-index: 200;
}
.card-tile:hover {
  transform: translateY(-4px);
  box-shadow: 0 16px 32px rgba(17, 10, 9, 0.36), 0 0 0 1px rgba(var(--vedh-primary-rgb), 0.12);
  border-color: rgba(var(--vedh-primary-rgb), 0.28);
}

.card-tile:hover .image {
  transform: scale(1.03);
  filter: saturate(1.04) contrast(1.02);
}

.selected {
  border-color: rgba(var(--vedh-secondary-rgb), 0.55);
  box-shadow: 0 0 0 1px rgba(var(--vedh-secondary-rgb), 0.2), 0 12px 28px rgba(21, 12, 9, 0.22);
}

.highlight {
  border-color: rgba(var(--vedh-primary-rgb), 0.58);
  box-shadow: 0 0 0 1px rgba(var(--vedh-primary-rgb), 0.22), 0 0 24px rgba(var(--vedh-primary-rgb), 0.16);
}

/* sizes: adjust text and padding, image will fill width of grid cell */
.size-xs { padding: 0.25rem; }
.size-sm { padding: 0.35rem; }
.size-md { padding: 0.4rem; }
.size-lg { padding: 0.5rem; }

.size-xs .label,
.size-sm .label {
  font-size: 0.72rem;
}

.size-lg .label {
  font-size: 0.86rem;
}
</style>
