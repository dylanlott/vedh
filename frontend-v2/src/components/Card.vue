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
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.4rem;
  cursor: grab;
  position: relative;
  transition: transform 140ms ease, box-shadow 160ms ease;
}
.card-tile:active { cursor: grabbing; }

.card-tile .label { font-size: 0.8rem; opacity: 0.9; }

.media {
  position: relative;
}
.media.image .image,
.media.thumb .image {
  width: 100%;
  aspect-ratio: 0.714;
  object-fit: cover;
  border-radius: 6px;
}
.media .image.placeholder {
  background: linear-gradient(135deg, rgba(40,40,40,0.9), rgba(22,22,22,0.9));
  color: rgba(255,255,255,0.7);
  display: grid;
  place-items: center;
}

/* tapped rotates the card image only */
.tapped .media .image { transform: rotate(90deg); }

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
  box-shadow: 0 8px 22px rgba(0,0,0,0.55);
}

/* sizes: adjust text and padding, image will fill width of grid cell */
.size-xs { padding: 0.25rem; }
.size-sm { padding: 0.35rem; }
.size-md { padding: 0.4rem; }
.size-lg { padding: 0.5rem; }
</style>