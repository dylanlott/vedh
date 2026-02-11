<template>
  <section class="board" v-if="game" :style="{ '--main-player-height': mainPlayerHeightCss }">
    <header class="board-header">
      <div class="title">
        <div class="turn-spotlight" :class="{ 'priority-owner': priorityOnTurn }">
          <div class="turn-number">
            <span class="eyebrow">Turn</span>
            <strong>{{ game.Turn?.Number ?? '—' }}</strong>
          </div>
          <div class="turn-phase">
            <span class="eyebrow">Phase</span>
            <strong>{{ game.Turn?.Phase ?? '—' }}</strong>
          </div>
          <div class="turn-priority">
            <span class="eyebrow">Priority</span>
            <strong>{{ currentPriority ?? '—' }}</strong>
          </div>
        </div>
        <p class="turn-meta">
          Active player: {{ game.Turn?.Player ?? 'Unknown player' }}
        </p>
        <div class="turn-tracker">
          <span class="chip">Turn {{ game.Turn?.Number ?? '—' }}</span>
          <span class="chip">Phase: {{ game.Turn?.Phase ?? '—' }}</span>
          <div class="turn-controls">
            <button class="tool" type="button" :disabled="!hasPriority || !nextPriorityPlayer" @click="passPriority">
              Pass priority
            </button>
            <button class="tool" type="button" :disabled="!isTurnPlayer || !nextPhase" @click="advancePhase(nextPhase)">
              Advance to {{ nextPhaseLabel }}
            </button>
            <button class="tool" type="button" :disabled="!hasPriority" @click="claimWin">
              Claim win
            </button>
          </div>
          <div v-if="pendingWinClaim" class="win-claim">
            <span>Win claim: {{ pendingWinClaim.ClaimedBy }}</span>
            <span class="muted">{{ pendingWinText }}</span>
          </div>
        </div>
      </div>
      <div class="settings">
        <button class="settings-trigger" type="button" @click="settingsOpen = !settingsOpen" aria-label="Open settings">
          ⚙️
        </button>
        <div v-if="settingsOpen" class="settings-drawer" @click.stop>
          <div class="settings-title">Settings</div>
          <label class="toast-toggle">
            <input type="checkbox" v-model="boardstateToastsEnabled" />
            <span>Board toasts</span>
          </label>
          <label class="toast-ttl">
            <span>Toast TTL</span>
            <input type="number" min="500" max="10000" step="250" v-model.number="toastDurationMs" />
            <span class="unit">ms</span>
          </label>
        </div>
      </div>
    </header>

    <div class="board-grid">
      <!-- Opponents at top -->
      <aside class="players opponents">
        <article v-for="player in opponents" :key="player.ID" :class="{ active: isActivePlayer(player.Username) }">
          <header>
            <h2>{{ player.Username }}</h2>
            <span class="life">{{ player.Boardstate?.Life ?? '—' }} life</span>
          </header>
          <div class="zone" :data-zone="'Commander'" :class="{ 'drag-over': isDragOver(player.Username, 'Commander') }" @dragenter.prevent="onDragEnter(player.Username, 'Commander')" @dragleave.prevent="onDragLeave(player.Username, 'Commander')">
              <h3>
                Commander
                <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Commander')">
                  {{ isStacked(player.Username, 'Commander') ? 'Tiles' : 'Art' }}
                </button>
              </h3>
              <template v-if="!isStacked(player.Username, 'Commander')">
                <ul class="cards tiles">
                  <li v-for="card in player.Boardstate?.Commander ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                    <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                    <span class="label">{{ card.Name }}</span>
                  </li>
                </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <li v-for="card in player.Boardstate?.Commander ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Battlefield'" :class="{ 'drag-over': isDragOver(player.Username, 'Battlefield') }" @dragenter.prevent="onDragEnter(player.Username, 'Battlefield')" @dragleave.prevent="onDragLeave(player.Username, 'Battlefield')">
              <h3>
                Battlefield
                <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Battlefield')">
                  {{ isStacked(player.Username, 'Battlefield') ? 'Tiles' : 'Art' }}
                </button>
              </h3>
              <template v-if="!isStacked(player.Username, 'Battlefield')">
                <ul class="cards tiles">
                  <li v-for="card in player.Boardstate?.Battlefield ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                    <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                    <span class="label">{{ card.Name }}</span>
                  </li>
                </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <li v-for="card in player.Boardstate?.Battlefield ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Hand'">
            <h3>Hand ({{ player.Boardstate?.Hand?.length ?? 0 }})</h3>
            <ul class="hand">
              <li class="card muted">Hidden</li>
            </ul>
          </div>
          <div class="zone" :data-zone="'Graveyard'" :class="{ 'drag-over': isDragOver(player.Username, 'Graveyard') }" @dragenter.prevent="onDragEnter(player.Username, 'Graveyard')" @dragleave.prevent="onDragLeave(player.Username, 'Graveyard')">
            <h3>
              Graveyard ({{ player.Boardstate?.Graveyard?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Graveyard')">
                {{ isStacked(player.Username, 'Graveyard') ? 'Tiles' : 'Art' }}
              </button>
            </h3>
            <template v-if="!isStacked(player.Username, 'Graveyard')">
              <ul class="cards tiles">
                <li v-for="card in player.Boardstate?.Graveyard ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                  <span class="label">{{ card.Name }}</span>
                </li>
              </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <li v-for="card in player.Boardstate?.Graveyard ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Exiled'" :class="{ 'drag-over': isDragOver(player.Username, 'Exiled') }" @dragenter.prevent="onDragEnter(player.Username, 'Exiled')" @dragleave.prevent="onDragLeave(player.Username, 'Exiled')">
            <h3>
              Exiled ({{ player.Boardstate?.Exiled?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Exiled')">
                {{ isStacked(player.Username, 'Exiled') ? 'Tiles' : 'Art' }}
              </button>
            </h3>
            <template v-if="!isStacked(player.Username, 'Exiled')">
              <ul class="cards tiles">
                <li v-for="card in player.Boardstate?.Exiled ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                  <span class="label">{{ card.Name }}</span>
                </li>
              </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <li v-for="card in player.Boardstate?.Exiled ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Revealed'" :class="{ 'drag-over': isDragOver(player.Username, 'Revealed') }" @dragenter.prevent="onDragEnter(player.Username, 'Revealed')" @dragleave.prevent="onDragLeave(player.Username, 'Revealed')">
            <h3>
              Revealed ({{ player.Boardstate?.Revealed?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Revealed')">
                {{ isStacked(player.Username, 'Revealed') ? 'Tiles' : 'Art' }}
              </button>
            </h3>
            <template v-if="!isStacked(player.Username, 'Revealed')">
              <ul class="cards tiles">
                <li v-for="card in player.Boardstate?.Revealed ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                  <span class="label">{{ card.Name }}</span>
                </li>
              </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <li v-for="card in player.Boardstate?.Revealed ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Controlled'" :class="{ 'drag-over': isDragOver(player.Username, 'Controlled') }" @dragenter.prevent="onDragEnter(player.Username, 'Controlled')" @dragleave.prevent="onDragLeave(player.Username, 'Controlled')">
            <h3>
              Controlled ({{ player.Boardstate?.Controlled?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Controlled')">
                {{ isStacked(player.Username, 'Controlled') ? 'Tiles' : 'Art' }}
              </button>
            </h3>
            <template v-if="!isStacked(player.Username, 'Controlled')">
              <ul class="cards tiles">
                <li v-for="card in player.Boardstate?.Controlled ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                  <span class="label">{{ card.Name }}</span>
                </li>
              </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <li v-for="card in player.Boardstate?.Controlled ?? []" :key="card.ID" :class="['card-tile', { dragging: currentDraggedId === card.ID }]">
                  <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Library'">
            <h3>Library ({{ player.Boardstate?.Library?.length ?? 0 }})</h3>
            <ul class="library">
              <li class="card muted">Hidden</li>
            </ul>
          </div>
        </article>
      </aside>
      <!-- self player is rendered below in the anchored .main-player section -->

      <!-- Stack separates opponents from self -->
      <section class="stack" :class="{ pulse: stackPulse }">
        <header>
          <h2>Stack</h2>
        </header>
        <ul class="cards tiles">
          <li v-for="(card, i) in game.Stack" :key="`${card.ID}-${i}`" class="card-tile stack-card">
            <button class="stack-resolve" type="button" @click.stop="resolveStackCard(i)">Resolve</button>
            <img :src="getImage(card.Name)" :alt="card.Name" @error="onImgError(card.Name)" />
            <span class="label">{{ card.Name }}</span>
          </li>
        </ul>
      </section>

      <!-- main-player moved outside of the scrollable .board-grid below -->

      
    </div>

    <!-- Main: current player's board (anchored to bottom) -->
    <section class="main-player" v-if="selfPlayer && !mainPlayerCollapsed">
      <div class="main-player-resize-handle" @dblclick="toggleMainPlayerCollapsed" @mousedown="startMainPlayerResize" />
      <article :class="{ active: isActivePlayer(selfPlayer.Username) }">
        <div class="main-player-left">
          <header class="main-player-header">
            <h2>{{ selfPlayer.Username }}</h2>
            <div class="life-row">
              <span class="life">{{ selfPlayer.Boardstate?.Life ?? '—' }} life</span>
              <div class="life-tools inline">
                <button class="tool" title="Lose 1 life" @click="changeLife(selfPlayer.Username, -1)">−1</button>
                <button class="tool" title="Gain 1 life" @click="changeLife(selfPlayer.Username, 1)">+1</button>
              </div>
            </div>
            <nav class="player-toolbar">
              <button
                class="tool"
                title="Draw 1"
                :disabled="selfLibraryKnown && selfLibraryCount === 0"
                @click="draw(selfPlayer.Username)"
              >🎴 Draw</button>
              <button
                class="tool"
                title="Mill 1"
                :disabled="selfLibraryKnown && selfLibraryCount === 0"
                @click="mill(selfPlayer.Username)"
              >🗑️ Mill</button>
              <button
                class="tool"
                title="Reveal top of library"
                :disabled="selfLibraryKnown && selfLibraryCount === 0"
                @click="revealTop(selfPlayer.Username)"
              >👁️ Reveal top</button>
              <button
                class="tool"
                title="Scry 1"
                :disabled="selfLibraryKnown && selfLibraryCount === 0"
                @click="scryOne(selfPlayer.Username)"
              >🔮 Scry 1</button>
              <button
                class="tool"
                title="Shuffle library"
                :disabled="selfLibraryKnown && (selfLibraryCount ?? 0) < 2"
                @click="shuffleLibrary(selfPlayer.Username)"
              >🔀 Shuffle</button>
            </nav>
          </header>
          <div class="zone commander" :data-zone="'Commander'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Commander') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Commander')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Commander')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Commander')">
            <h3>
              Commander
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Commander')">
              {{ isStacked(selfPlayer.Username, 'Commander') ? 'Tiles' : 'Art' }}
              </button>
            </h3>
            <template v-if="!isStacked(selfPlayer.Username, 'Commander')">
              <ul class="cards tiles">
                <Card
                  v-for="(card, idx) in selfPlayer.Boardstate?.Commander ?? []"
                  :key="card.ID || `${card.Name}-${idx}`"
                  :id="card.ID"
                  :name="card.Name"
                  :image-src="getImage(card.Name)"
                  :tapped="isCardTapped(card)"
                  :draggable="true"
                  :dragging="currentDraggedId === card.ID"
                  @dragstart="() => onDragStart(card, me.Username, 'Commander', idx)"
                  @dragend="onDragEnd"
                  @click="() => onCardClick(card, me.Username, 'Commander', idx)"
                  @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Commander', idx)"
                />
              </ul>
            </template>
            <template v-else>
              <ul class="cards art">
                <Card
                  v-for="(card, idx) in selfPlayer.Boardstate?.Commander ?? []"
                  :key="card.ID || `${card.Name}-${idx}`"
                  :id="card.ID"
                  :name="card.Name"
                  :image-src="getImage(card.Name)"
                  :tapped="isCardTapped(card)"
                  :draggable="true"
                  :dragging="currentDraggedId === card.ID"
                  :show-name="false"
                  size="lg"
                  @dragstart="() => onDragStart(card, me.Username, 'Commander', idx)"
                  @dragend="onDragEnd"
                  @click="() => onCardClick(card, me.Username, 'Commander', idx)"
                  @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Commander', idx)"
                />
              </ul>
            </template>
          </div>
        </div>
        <div class="main-player-right">
  <div class="zone" :data-zone="'Battlefield'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Battlefield') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Battlefield')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Battlefield')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Battlefield')">
          <h3>
            Battlefield
            <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Battlefield')">
              {{ isStacked(selfPlayer.Username, 'Battlefield') ? 'Tiles' : 'Art' }}
            </button>
          </h3>
          <template v-if="!isStacked(selfPlayer.Username, 'Battlefield')">
            <ul class="cards tiles">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Battlefield ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                @dragstart="() => onDragStart(card, me.Username, 'Battlefield', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Battlefield', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Battlefield', idx)"
              />
            </ul>
          </template>
          <template v-else>
            <ul class="cards art">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Battlefield ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                :show-name="false"
                size="lg"
                @dragstart="() => onDragStart(card, me.Username, 'Battlefield', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Battlefield', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Battlefield', idx)"
              />
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Hand'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Hand') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Hand')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Hand')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Hand')">
          <h3>
            Hand ({{ selfPlayer.Boardstate?.Hand?.length ?? 0 }})
            <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Hand')">
              {{ isStacked(selfPlayer.Username, 'Hand') ? 'Tiles' : 'Art' }}
            </button>
          </h3>
          <template v-if="!isStacked(selfPlayer.Username, 'Hand')">
            <ul class="cards tiles">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Hand ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                @dragstart="() => onDragStart(card, me.Username, 'Hand', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Hand', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Hand', idx)"
              />
            </ul>
          </template>
          <template v-else>
            <ul class="cards art">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Hand ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                :show-name="false"
                size="lg"
                @dragstart="() => onDragStart(card, me.Username, 'Hand', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Hand', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Hand', idx)"
              />
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Graveyard'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Graveyard') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Graveyard')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Graveyard')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Graveyard')">
          <h3>
            Graveyard ({{ selfPlayer.Boardstate?.Graveyard?.length ?? 0 }})
            <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Graveyard')">
              {{ isStacked(selfPlayer.Username, 'Graveyard') ? 'Tiles' : 'Art' }}
            </button>
          </h3>
          <template v-if="!isStacked(selfPlayer.Username, 'Graveyard')">
            <ul class="cards tiles">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Graveyard ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                @dragstart="() => onDragStart(card, me.Username, 'Graveyard', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Graveyard', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Graveyard', idx)"
              />
            </ul>
          </template>
          <template v-else>
            <ul class="cards art">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Graveyard ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                :show-name="false"
                size="lg"
                @dragstart="() => onDragStart(card, me.Username, 'Graveyard', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Graveyard', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Graveyard', idx)"
              />
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Exiled'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Exiled') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Exiled')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Exiled')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Exiled')">
          <h3>
            Exiled ({{ selfPlayer.Boardstate?.Exiled?.length ?? 0 }})
            <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Exiled')">
              {{ isStacked(selfPlayer.Username, 'Exiled') ? 'Tiles' : 'Art' }}
            </button>
          </h3>
          <template v-if="!isStacked(selfPlayer.Username, 'Exiled')">
            <ul class="cards tiles">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Exiled ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                @dragstart="() => onDragStart(card, me.Username, 'Exiled', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Exiled', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Exiled', idx)"
              />
            </ul>
          </template>
          <template v-else>
            <ul class="cards art">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Exiled ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                :show-name="false"
                size="lg"
                @dragstart="() => onDragStart(card, me.Username, 'Exiled', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Exiled', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Exiled', idx)"
              />
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Revealed'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Revealed') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Revealed')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Revealed')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Revealed')">
          <h3>
            Revealed ({{ selfPlayer.Boardstate?.Revealed?.length ?? 0 }})
            <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Revealed')">
              {{ isStacked(selfPlayer.Username, 'Revealed') ? 'Tiles' : 'Art' }}
            </button>
          </h3>
          <template v-if="!isStacked(selfPlayer.Username, 'Revealed')">
            <ul class="cards tiles">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Revealed ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                @dragstart="() => onDragStart(card, me.Username, 'Revealed', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Revealed', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Revealed', idx)"
              />
            </ul>
          </template>
          <template v-else>
            <ul class="cards art">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Revealed ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                :show-name="false"
                size="lg"
                @dragstart="() => onDragStart(card, me.Username, 'Revealed', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Revealed', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Revealed', idx)"
              />
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Controlled'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Controlled') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Controlled')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Controlled')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Controlled')">
          <h3>
            Controlled ({{ selfPlayer.Boardstate?.Controlled?.length ?? 0 }})
            <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(selfPlayer.Username, 'Controlled')">
              {{ isStacked(selfPlayer.Username, 'Controlled') ? 'Tiles' : 'Art' }}
            </button>
          </h3>
          <template v-if="!isStacked(selfPlayer.Username, 'Controlled')">
            <ul class="cards tiles">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Controlled ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                @dragstart="() => onDragStart(card, me.Username, 'Controlled', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Controlled', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Controlled', idx)"
              />
            </ul>
          </template>
          <template v-else>
            <ul class="cards art">
              <Card
                v-for="(card, idx) in selfPlayer.Boardstate?.Controlled ?? []"
                :key="card.ID || `${card.Name}-${idx}`"
                :id="card.ID"
                :name="card.Name"
                :image-src="getImage(card.Name)"
                :tapped="isCardTapped(card)"
                :draggable="true"
                :dragging="currentDraggedId === card.ID"
                :show-name="false"
                size="lg"
                @dragstart="() => onDragStart(card, me.Username, 'Controlled', idx)"
                @dragend="onDragEnd"
                @click="() => onCardClick(card, me.Username, 'Controlled', idx)"
                @dblclick.stop.prevent="() => onCardDblClick(card, me.Username, 'Controlled', idx)"
              />
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Library'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Library') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Library')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Library')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Library')">
          <h3>
            Library ({{ selfLibraryKnown ? selfLibraryCount : '—' }})
          </h3>
          <ul class="library">
            <li v-if="selfLibraryKnown && (selfLibraryCount ?? 0) === 0" class="card muted">Empty</li>
          </ul>
        </div>
        </div>
      </article>
    </section>
    <button
      v-if="selfPlayer && mainPlayerCollapsed"
      class="main-player-collapsed-toggle"
      type="button"
      title="Show player panel"
      @click="toggleMainPlayerCollapsed"
    >
      ▲
    </button>
  </section>
  <section v-else class="loading-state">
    <p>Loading game…</p>
  </section>

  <!-- Scry 1 modal (self-only) -->
  <div v-if="scry?.open && isSelf(scry?.username)" class="scry-overlay">
    <div class="scry-modal">
      <header>Scry 1</header>
      <img
        v-if="scry?.card"
        class="scry-card"
        :src="getImage(scry.card.Name)"
        :alt="scry.card.Name"
        @error="onImgError(scry.card.Name)"
      />
      <p v-if="scry?.card">Top card: <strong>{{ scry.card.Name }}</strong></p>
      <div class="scry-actions">
        <button class="tool" @click="scryKeepTop">Keep on top</button>
        <button class="tool" @click="scryPutBottom">Put on bottom</button>
      </div>
    </div>
  </div>

  <!-- Toasts -->
  <div class="toasts" :class="{ compact: mainPlayerCollapsed }">
    <div class="toast" v-for="t in toasts" :key="t.id">{{ t.text }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useGamesStore } from '../stores/games';
import { useAuthStore } from '../stores/auth';
import { apolloClient } from '../services/apollo';
import { ADVANCE_PHASE_MUTATION, CLAIM_WIN_MUTATION, PASS_PRIORITY_MUTATION, UPDATE_BOARDSTATE_MUTATION, UPDATE_GAME_MUTATION } from '../graphql/mutations';
// Subscriptions are handled centrally in the games store.
import { fetchScryfallImageByName } from '../services/scryfall';
import Card from '../components/Card.vue';
import { isLandCard, moveHandCardToStackState, resolveStackCardToGraveyardState } from '../utils/stack';
// Dev logging helper: use console.log so messages appear without enabling Verbose level
function dbg(...args: any[]) { console.log(...args); }

const games = useGamesStore();
const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

// Zone typing shared across helpers
const zones = ['Commander','Battlefield','Hand','Graveyard','Exiled','Revealed','Library','Controlled'] as const;
type Zone = typeof zones[number];

const game = computed(() => games.activeGame);

// Current user player object and opponents
const selfPlayer = computed(() => {
  const username = auth.profile?.Username;
  return game.value?.Players.find(p => p.Username === username) ?? null;
});
const opponents = computed(() => game.value?.Players.filter(p => p.Username !== auth.profile?.Username) ?? []);
// Non-null alias for template usage (we guard rendering with v-if="selfPlayer")
const me = computed(() => selfPlayer.value!);
const selfLibraryCount = computed(() => selfPlayer.value?.Boardstate?.Library?.length ?? null);
const selfLibraryKnown = computed(() => Array.isArray(selfPlayer.value?.Boardstate?.Library));
const currentPriority = computed(() => game.value?.Turn?.Priority ?? game.value?.Turn?.Player ?? null);
const hasPriority = computed(() => {
  if (!currentPriority.value) return true;
  return auth.profile?.Username === currentPriority.value;
});
const isTurnPlayer = computed(() => auth.profile?.Username === game.value?.Turn?.Player);
const priorityOnTurn = computed(() => {
  const turnPlayer = game.value?.Turn?.Player ?? null;
  return !!turnPlayer && currentPriority.value === turnPlayer;
});
const nextPriorityPlayer = computed(() => {
  const players = (game.value?.Players ?? []).map(p => p.Username).filter(Boolean);
  if (!players.length) return '';
  const holder = currentPriority.value ?? game.value?.Turn?.Player ?? '';
  const holderIdx = players.findIndex(name => name === holder);
  if (holderIdx === -1) return players[0] ?? '';
  return players[(holderIdx + 1) % players.length] ?? '';
});
const pendingWinClaim = computed(() => game.value?.PendingWinClaim ?? null);
const pendingWinText = computed(() => {
  if (!pendingWinClaim.value) return '';
  const next = pendingWinClaim.value.Remaining?.[0];
  return next ? `Awaiting ${next}` : 'Awaiting priority';
});

// Simple tile-only view; no display toggles needed
const MAIN_PLAYER_PREF_KEY = 'vedh:mainPlayerPanel:v1';
const BOARDSTATE_TOASTS_KEY = 'vedh:boardstateToasts:v1';
const TOAST_TTL_KEY = 'vedh:toastTtlMs:v1';
const boardstateToastsEnabled = ref(true);
const toastDurationMs = ref(2500);
const settingsOpen = ref(false);
const stackPulse = ref(false);
let stackPulseTimer: number | null = null;
const boardstateToastReady = ref(false);
const lastBoardstateSnapshots = ref<Record<string, { life: number; zones: Record<Zone, { items: { id: string; name: string }[] }> }>>({});
const mainPlayerHeight = ref(33);
const mainPlayerCollapsed = ref(false);
const mainPlayerHeightCss = computed(() => mainPlayerCollapsed.value ? '0px' : `${mainPlayerHeight.value}vh`);
const mainPlayerResizing = ref(false);
const mainPlayerResizeStart = ref({ y: 0, height: 33 });
const MAIN_PLAYER_MIN_VH = 16;
const MAIN_PLAYER_MAX_VH = 60;
const PHASE_ORDER = [
  { key: 'UNTAP', label: 'Untap' },
  { key: 'UPKEEP', label: 'Upkeep' },
  { key: 'DRAW', label: 'Draw' },
  { key: 'MAIN PHASE 1', label: 'Main Phase 1' },
  { key: 'COMBAT', label: 'Combat' },
  { key: 'MAIN PHASE 2', label: 'Main Phase 2' },
  { key: 'END STEP', label: 'End Step' },
  { key: 'DISCARD', label: 'Discard' },
];

const normalizedPhaseKey = (phase: string | null | undefined) => {
  if (!phase) return null;
  const cleaned = phase.toUpperCase().replace(/\s+/g, ' ').trim();
  switch (cleaned) {
    case 'MAIN':
    case 'MAIN PHASE':
    case 'MAIN1':
    case 'MAIN PHASE 1':
      return 'MAIN PHASE 1';
    case 'MAIN2':
    case 'MAIN PHASE 2':
      return 'MAIN PHASE 2';
    case 'END':
    case 'END STEP':
      return 'END STEP';
    default:
      return cleaned;
  }
};

const nextPhase = computed(() => {
  const currentKey = normalizedPhaseKey(game.value?.Turn?.Phase);
  const idx = PHASE_ORDER.findIndex(p => p.key === currentKey);
  const next = PHASE_ORDER[idx + 1] ?? PHASE_ORDER[0];
  return next?.key ?? null;
});

const nextPhaseLabel = computed(() => {
  const currentKey = normalizedPhaseKey(game.value?.Turn?.Phase);
  const idx = PHASE_ORDER.findIndex(p => p.key === currentKey);
  const next = PHASE_ORDER[idx + 1] ?? PHASE_ORDER[0];
  return next?.label ?? 'Next phase';
});

onMounted(() => {
  try {
    const raw = localStorage.getItem(MAIN_PLAYER_PREF_KEY);
    if (!raw) return;
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === 'object') {
      if (typeof parsed.height === 'number') {
        mainPlayerHeight.value = Math.min(60, Math.max(16, parsed.height));
      }
      if (typeof parsed.collapsed === 'boolean') {
        mainPlayerCollapsed.value = parsed.collapsed;
      }
    }
  } catch {}
  try {
    const rawToast = localStorage.getItem(BOARDSTATE_TOASTS_KEY);
    if (rawToast !== null) {
      boardstateToastsEnabled.value = rawToast === 'true';
    }
  } catch {}
  try {
    const rawTtl = localStorage.getItem(TOAST_TTL_KEY);
    if (rawTtl) {
      const parsed = Number(rawTtl);
      if (!Number.isNaN(parsed)) {
        toastDurationMs.value = Math.min(10000, Math.max(500, parsed));
      }
    }
  } catch {}
});
watch([mainPlayerHeight, mainPlayerCollapsed], () => {
  try {
    localStorage.setItem(MAIN_PLAYER_PREF_KEY, JSON.stringify({
      height: mainPlayerHeight.value,
      collapsed: mainPlayerCollapsed.value,
    }));
  } catch {}
});
watch(boardstateToastsEnabled, (val) => {
  try {
    localStorage.setItem(BOARDSTATE_TOASTS_KEY, String(val));
  } catch {}
});
watch(toastDurationMs, (val) => {
  if (Number.isNaN(val)) return;
  const clamped = Math.min(10000, Math.max(500, val));
  if (clamped !== val) {
    toastDurationMs.value = clamped;
    return;
  }
  try {
    localStorage.setItem(TOAST_TTL_KEY, String(clamped));
  } catch {}
});

function toggleMainPlayerCollapsed() {
  mainPlayerCollapsed.value = !mainPlayerCollapsed.value;
}

function onMainPlayerResizeMove(event: MouseEvent) {
  if (!mainPlayerResizing.value) return;
  const viewportHeight = window.innerHeight || 1;
  const delta = mainPlayerResizeStart.value.y - event.clientY;
  const deltaVh = (delta / viewportHeight) * 100;
  const next = Math.min(
    MAIN_PLAYER_MAX_VH,
    Math.max(MAIN_PLAYER_MIN_VH, mainPlayerResizeStart.value.height + deltaVh),
  );
  mainPlayerHeight.value = Math.round(next);
}

async function passPriority() {
  const g = game.value;
  const toPlayer = nextPriorityPlayer.value;
  if (!g || !toPlayer) return;
  try {
    await apolloClient.mutate({
      mutation: PASS_PRIORITY_MUTATION,
      variables: { gameID: g.ID, toPlayer },
    });
    addToast(`Passed priority to ${toPlayer}`);
  } catch (e) {
    console.error('[board] passPriority failed', e);
    addToast('Failed to pass priority');
  }
}

async function advancePhase(phase: string) {
  const g = game.value;
  if (!g || !phase) return;
  try {
    await apolloClient.mutate({
      mutation: ADVANCE_PHASE_MUTATION,
      variables: { gameID: g.ID, phase },
    });
    addToast(`Advanced phase to ${phase}`);
  } catch (e) {
    console.error('[board] advancePhase failed', e);
    addToast('Failed to advance phase');
  }
}

async function claimWin() {
  const g = game.value;
  if (!g) return;
  const response = window.prompt('Win condition (optional)', '');
  if (response === null) return;
  const condition = response.trim();
  try {
    await apolloClient.mutate({
      mutation: CLAIM_WIN_MUTATION,
      variables: { gameID: g.ID, condition: condition || null },
    });
    addToast('Win claimed');
  } catch (e) {
    console.error('[board] claimWin failed', e);
    addToast('Failed to claim win');
  }
}

function stopMainPlayerResize() {
  if (!mainPlayerResizing.value) return;
  mainPlayerResizing.value = false;
  document.removeEventListener('mousemove', onMainPlayerResizeMove);
  document.removeEventListener('mouseup', stopMainPlayerResize);
}

function startMainPlayerResize(event: MouseEvent) {
  if (mainPlayerCollapsed.value) return;
  mainPlayerResizing.value = true;
  mainPlayerResizeStart.value = { y: event.clientY, height: mainPlayerHeight.value };
  document.addEventListener('mousemove', onMainPlayerResizeMove);
  document.addEventListener('mouseup', stopMainPlayerResize);
}

const PUBLIC_ZONES: Zone[] = ['Commander', 'Battlefield', 'Graveyard', 'Exiled', 'Revealed', 'Controlled'];
const PRIVATE_ZONES: Zone[] = ['Hand', 'Library'];

function snapshotBoardstate(player: { Username?: string; Boardstate?: any }) {
  const bs = player.Boardstate ?? {};
  const zonesSnap: Record<Zone, { items: { id: string; name: string }[] }> = {} as any;
  for (const z of zones) {
    const list = Array.isArray(bs[z]) ? bs[z] : [];
    zonesSnap[z] = {
      items: list.map((c: any) => ({ id: c?.ID ?? '', name: c?.Name ?? '' })),
    };
  }
  return {
    life: bs.Life ?? 0,
    zones: zonesSnap,
  };
}

function countByKey(items: { id: string; name: string }[]) {
  const counts = new Map<string, { count: number; name: string }>();
  for (const item of items) {
    const key = item.id || item.name || '';
    if (!key) continue;
    const prev = counts.get(key);
    if (prev) {
      prev.count += 1;
    } else {
      counts.set(key, { count: 1, name: item.name || key });
    }
  }
  return counts;
}

function describeZoneDelta(
  username: string,
  zone: Zone,
  prevItems: { id: string; name: string }[],
  nextItems: { id: string; name: string }[],
) {
  const prevCount = prevItems.length;
  const nextCount = nextItems.length;
  if (prevCount === nextCount) return [] as string[];
  const delta = nextCount - prevCount;
  const isPublic = PUBLIC_ZONES.includes(zone);
  if (!isPublic) {
    if (zone === 'Hand' && delta === 1) return [`${username} drew a card`];
    return [`${username} ${delta > 0 ? 'added' : 'removed'} ${Math.abs(delta)} card${Math.abs(delta) === 1 ? '' : 's'} ${delta > 0 ? 'to' : 'from'} ${zone}`];
  }
  const prevCounts = countByKey(prevItems);
  const nextCounts = countByKey(nextItems);
  const added: { name: string }[] = [];
  const removed: { name: string }[] = [];
  for (const [key, { count, name }] of nextCounts.entries()) {
    const before = prevCounts.get(key)?.count ?? 0;
    if (count > before) {
      for (let i = 0; i < count - before; i++) added.push({ name });
    }
  }
  for (const [key, { count, name }] of prevCounts.entries()) {
    const after = nextCounts.get(key)?.count ?? 0;
    if (count > after) {
      for (let i = 0; i < count - after; i++) removed.push({ name });
    }
  }
  if (added.length === 1 && delta === 1) {
    return [`${username} put ${added[0].name || 'a card'} into ${zone}`];
  }
  if (removed.length === 1 && delta === -1) {
    return [`${username} removed ${removed[0].name || 'a card'} from ${zone}`];
  }
  return [`${username} ${delta > 0 ? 'added' : 'removed'} ${Math.abs(delta)} card${Math.abs(delta) === 1 ? '' : 's'} ${delta > 0 ? 'to' : 'from'} ${zone}`];
}

watch(() => game.value?.Players, (players) => {
  if (!players) return;
  const next: Record<string, { life: number; zones: Record<Zone, { items: { id: string; name: string }[] }> }> = {};
  for (const p of players) {
    const key = p.Username ?? p.ID ?? 'unknown';
    next[key] = snapshotBoardstate(p);
  }
  if (boardstateToastReady.value && boardstateToastsEnabled.value) {
    for (const [user, snap] of Object.entries(next)) {
      const prev = lastBoardstateSnapshots.value[user];
      if (!prev) continue;
      if (snap.life !== prev.life) {
        const delta = snap.life - prev.life;
        addToast(`${user} ${delta >= 0 ? 'gained' : 'lost'} ${Math.abs(delta)} life`, toastDurationMs.value);
      }
      for (const z of zones) {
        const prevItems = prev.zones[z]?.items ?? [];
        const nextItems = snap.zones[z]?.items ?? [];
        const messages = describeZoneDelta(user, z, prevItems, nextItems);
        for (const msg of messages) addToast(msg, toastDurationMs.value);
      }
    }
  }
  lastBoardstateSnapshots.value = next;
  if (!boardstateToastReady.value) {
    boardstateToastReady.value = true;
  }
}, { deep: true });

// Image cache and helpers
const imageCache = ref<Record<string, string | null>>({});
const placeholder = 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="200" height="280"><rect width="100%" height="100%" fill="%23222"/><text x="50%" y="50%" dominant-baseline="middle" text-anchor="middle" fill="%23aaa" font-size="14">Loading…</text></svg>';
function onImgError(name: string) {
  // mark as null to avoid infinite error loops
  imageCache.value[name] = null;
}
async function ensureImage(name: string) {
  if (imageCache.value[name] !== undefined) return;
  imageCache.value[name] = null;
  const url = await fetchScryfallImageByName(name);
  imageCache.value[name] = url;
  dbg('[display] ensureImage', { name, url });
}
function uniqueNamesFrom(list: { Name: string }[]) {
  return Array.from(new Set(list.map(c => c.Name)));
}
function prefetchVisibleImages() {
  const g = game.value;
  if (!g) return;
  const names = new Set<string>();
  for (const p of g.Players) {
    const bs: any = p.Boardstate || {};
    for (const z of zones) {
      (bs[z] ?? []).forEach((c: any) => names.add(c.Name));
    }
  }
  // Include the global stack
  (g.Stack ?? []).forEach((c: any) => names.add(c.Name));
  const list = Array.from(names).slice(0, 150);
  dbg('[display] prefetchVisibleImages', { count: list.length, sample: list.slice(0, 5) });
  list.forEach(n => ensureImage(n));
}
watch(() => game.value?.ID, () => {
  prefetchVisibleImages();
});
watch(() => game.value?.Stack?.length, () => {
  // Prefetch when stack changes to keep tiles updated for all viewers
  prefetchVisibleImages();
});
watch(() => game.value?.Stack?.length, (next, prev) => {
  if (prev === undefined || next === undefined) return;
  if (next > prev) {
    stackPulse.value = false;
    if (stackPulseTimer) window.clearTimeout(stackPulseTimer);
    // retrigger animation
    stackPulseTimer = window.setTimeout(() => {
      stackPulse.value = true;
      stackPulseTimer = window.setTimeout(() => {
        stackPulse.value = false;
      }, 900);
    }, 0);
  }
});

watch(() => game.value?.Status, (status) => {
  if (status === 'FINISHED' && game.value?.ID) {
    router.push({ name: 'game-analysis', params: { id: game.value.ID } });
  }
});

// Lazy accessor for image src: kicks off fetch on first access
function getImage(name: string): string {
  if (!(name in imageCache.value)) {
    ensureImage(name);
  }
  return imageCache.value[name] || placeholder;
}

function buildBoardstateInput(
  player: any,
  userFallback: string,
  gameID: string,
  overrides: Partial<Record<Zone, any[]>> = {},
  lifeOverride?: number,
) {
  const input: any = {
    UserID: player.ID ?? userFallback,
    User: player.Username,
    GameID: gameID,
    Life: lifeOverride ?? player.Boardstate?.Life ?? 0,
  };
  for (const z of zones) {
    const list = overrides[z] ?? player.Boardstate?.[z as Zone];
    if (Array.isArray(list)) {
      input[z] = list;
    }
  }
  return input;
}

function buildGameInputFromGame(
  g: any,
  playersOverride?: any[],
  stackOverride?: any[],
) {
  const players = (playersOverride ?? g.Players ?? []).map((p: any) => ({
    ID: p.ID,
    Username: p.Username,
    Boardstate: p.Boardstate ? buildBoardstateInput(p, p.Username, g.ID) : undefined,
  }));
  return {
    ID: g.ID,
    CreatedAt: g.CreatedAt,
    Turn: g.Turn,
    Rules: g.Rules,
    Players: players,
    Stack: stackOverride ?? g.Stack ?? [],
  };
}

onMounted(async () => {
  const gameID = route.params.id as string;
  dbg('[display] mounted');
  await games.loadGame(gameID, auth.profile?.ID);
  prefetchVisibleImages();
});

// Direct local subscription: sometimes the store-level subscription may not
// surface immediately to this view (or may be replaced elsewhere). Create a
// local subscription here to ensure we always receive game updates and apply
// them to the `games` store so opponent boardstates update in the UI.
const localGameSubscription: { unsubscribe?: () => void } = {};
// Local subscription removed to avoid duplicating the store subscription and
// potentially replacing the same userID observer on the server. The store
// manages a single subscription per game/user.

// Ensure we (re)subscribe once auth/userID is known and route is set.
// This covers cases where the auth profile is populated after the view mounts
// so the initial subscription might have been attempted without a userID.
watch([
  () => auth.profile?.ID,
  () => route.params.id,
], ([userID, gameID]) => {
  if (typeof gameID === 'string' && gameID && typeof userID === 'string' && userID) {
    dbg('[display] ensure subscription', { gameID, userID });
    games.subscribeToGame(gameID, userID);
  }
});

onBeforeUnmount(() => {
  games.clearActiveGame();
  stopMainPlayerResize();
  if (stackPulseTimer) window.clearTimeout(stackPulseTimer);
});

function isActivePlayer(username: string) {
  return username === game.value?.Turn?.Player;
}

function isSelf(username: string) {
  return username === auth.profile?.Username;
}

// Basic drag-and-drop state
const dragged = ref<{ card: { ID: string; Name: string }; fromUser: string; fromZone: Zone; fromIndex?: number } | null>(null);
// id of the currently dragged card (for CSS/animations)
const currentDraggedId = ref<string | null>(null);
// drag-over tracking per zone (keyed by "username::zone")
const dragOver = ref<Record<string, boolean>>({});
function dragKey(user: string, zone: string) { return `${user}::${zone}`; }
function onDragEnter(user: string, zone: string) {
  if (!dragged.value) return;
  dragOver.value[dragKey(user, zone)] = true;
}
function onDragLeave(user: string, zone: string) {
  dragOver.value[dragKey(user, zone)] = false;
}
function isDragOver(user: string, zone: string) { return !!dragOver.value[dragKey(user, zone)]; }
function clearDragOverAll() { dragOver.value = {}; }

function onDragStart(card: { ID: string; Name: string }, fromUser: string, fromZone: string, fromIndex?: number) {
  dragged.value = { card, fromUser, fromZone: fromZone as Zone, fromIndex };
  currentDraggedId.value = card.ID;
}

function onDragEnd() {
  dragged.value = null;
  currentDraggedId.value = null;
  clearDragOverAll();
}

function onDrop(toUser: string, toZone: string) {
  return (async () => {
    if (!dragged.value || !game.value) return;
    await moveCard({
      gameID: game.value.ID,
      user: toUser,
      fromUser: dragged.value.fromUser,
      cardID: dragged.value.card.ID,
      fromZone: dragged.value.fromZone,
      toZone: toZone as Zone,
      fromIndex: dragged.value.fromIndex,
    });
    // clear drag state and animations
    dragged.value = null;
    currentDraggedId.value = null;
    dragOver.value[dragKey(toUser, toZone)] = false;
    clearDragOverAll();
  })();
}

// Simple click-to-move: toggles between Hand and Battlefield for demo
async function quickMove(card: { ID: string; Name: string }, user: string, zone: string, fromIndex?: number) {
  if (!game.value) return;
  const toZone: Zone = (zone === 'Hand' ? 'Battlefield' : 'Hand') as Zone;
  await moveCard({
    gameID: game.value.ID,
    user,
    fromUser: user,
    cardID: card.ID,
    fromZone: zone as Zone,
    toZone,
    fromIndex,
  });
}

type MoveCardArgs = {
  gameID: string;
  user: string;
  fromUser: string;
  cardID: string;
  fromZone: Zone;
  toZone: Zone;
  fromIndex?: number;
  toIndex?: number;
};

const PERSISTED_ZONES: Zone[] = ['Library', 'Hand', 'Graveyard', 'Revealed', 'Controlled'];

async function moveCard(args: MoveCardArgs) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === args.user);
  if (!player || !player.Boardstate) return;

  // Clone current zones only when present to avoid wiping unknown zones
  const current: Partial<Record<Zone, any[]>> = {};
  for (const z of zones) {
    const list = player.Boardstate?.[z as Zone];
    if (Array.isArray(list)) {
      current[z] = [...list];
    }
  }
  if (!current[args.fromZone]) current[args.fromZone] = [];
  if (!current[args.toZone]) current[args.toZone] = [];

  // Find full card details from source player's zones to preserve Name
  const sourcePlayer = g.Players.find(p => p.Username === args.fromUser);
  let movedCard: any | null = null;
  if (sourcePlayer?.Boardstate) {
    for (const z of zones) {
      const list = (sourcePlayer.Boardstate as any)[z] ?? [];
      let found: any | null = null;
      if (args.cardID) {
        found = list.find((c: any) => c.ID === args.cardID);
      }
      if (!found && typeof args.fromIndex === 'number' && z === args.fromZone) {
        found = list[args.fromIndex];
      }
      if (found) { movedCard = { ...found }; break; }
    }
  }

  // Remove from source zone (if same user). If Library cards don't have IDs, fall back to index 0.
  if (args.fromUser === args.user) {
    const sourceList = current[args.fromZone as Zone] ?? [];
    let removeIndex = -1;
    if (args.cardID) {
      removeIndex = sourceList.findIndex(c => c.ID === args.cardID);
    }
    if (removeIndex === -1 && typeof args.fromIndex === 'number') {
      removeIndex = args.fromIndex;
    }
    if (removeIndex === -1 && args.fromZone === 'Library' && sourceList.length > 0) {
      removeIndex = 0;
    }
    if (removeIndex !== -1 && removeIndex < sourceList.length) {
      current[args.fromZone as Zone] = sourceList.filter((_, idx) => idx !== removeIndex);
      if (!movedCard) {
        movedCard = { ...sourceList[removeIndex] };
      }
    } else if (args.cardID) {
      current[args.fromZone as Zone] = sourceList.filter(c => c.ID !== args.cardID);
    }
  }
  const isEphemeral = !(movedCard?.ID || args.cardID) && !PERSISTED_ZONES.includes(args.fromZone);
  if (args.cardID) {
    // Ensure uniqueness across zones for real cards
    for (const z of zones) {
      if (!current[z]) continue;
      current[z] = (current[z] ?? []).filter((c: any) => c.ID !== args.cardID);
    }
  }
  // Add to destination zone (dedupe by ID when available); tokens/clones vanish in persisted zones.
  const destList = current[args.toZone as Zone] ?? [];
  if (isEphemeral && PERSISTED_ZONES.includes(args.toZone)) {
    // token/clone leaves play and doesn't persist
  } else if (!args.cardID || !destList.some((c: any) => c.ID === args.cardID)) {
    const nextCard = movedCard ?? { ID: args.cardID ?? '', Name: '' };
    if (args.toZone !== 'Battlefield' && nextCard && 'Tapped' in nextCard) {
      nextCard.Tapped = false;
    }
    const targetIndex = typeof args.toIndex === 'number'
      ? Math.max(0, Math.min(args.toIndex, destList.length))
      : 0;
    destList.splice(targetIndex, 0, nextCard);
  }
  current[args.toZone as Zone] = destList;

  const input = buildBoardstateInput(player, args.user, g.ID, current, player.Boardstate.Life);

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    // Optimistically update local store
    applyLocalBoardstatePatch(args.user, draft => ({
      ...draft,
      ...current,
      Life: player.Boardstate!.Life,
    }));
    if (args.fromUser !== args.user) {
      // Remove card from source player locally (cross-player moves)
      applyLocalBoardstatePatch(args.fromUser, (draft: any) => ({
        ...draft,
        [args.fromZone]: (draft[args.fromZone] ?? []).filter((c: { ID: string }) => c.ID !== args.cardID),
      }));
    }
  } catch (e) {
    console.error('[board] moveCard failed', e);
  }
}

// Toggle tapped state for a specific card within a zone and persist
async function toggleTapped(user: string, zone: Zone, card: any, fromIndex?: number) {
  if (zone !== 'Battlefield') return;
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === user);
  if (!player || !player.Boardstate) return;

  // Clone zones only when present to avoid wiping unknown zones
  const current: Partial<Record<Zone, any[]>> = {};
  for (const z of zones) {
    const list = player.Boardstate?.[z as Zone];
    if (Array.isArray(list)) {
      current[z] = [...list];
    }
  }
  if (!current[zone]) current[zone] = [];

  // Toggle tapped on the matching card in the zone
  const list = current[zone] ?? [];
  const updated = list.map((c: any, idx: number) => {
    if (typeof fromIndex === 'number') {
      return idx === fromIndex ? { ...c, Tapped: !c?.Tapped } : c;
    }
    if (card.ID && c.ID === card.ID) {
      return { ...c, Tapped: !c?.Tapped };
    }
    return c;
  });
  current[zone] = updated;

  const input = buildBoardstateInput(player, user, g.ID, current, player.Boardstate.Life);

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    // Optimistic local update
    applyLocalBoardstatePatch(user, (draft: any) => ({
      ...draft,
      ...current,
      Life: player.Boardstate!.Life,
    }));
  } catch (e) {
    console.error('[board] toggleTapped failed', e);
  }
}

// Helper for template typing
function isCardTapped(card: any): boolean { return !!card?.Tapped; }

async function moveHandCardToStack(card: any, user: string, fromIndex?: number) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === user);
  if (!player?.Boardstate?.Hand) return;

  const moved = moveHandCardToStackState(
    player.Boardstate.Hand,
    g.Stack ?? [],
    card,
    user,
    fromIndex,
  );
  if (moved.skippedReason === 'duplicate') {
    addToast('Card is already on the stack');
    return;
  }
  if (moved.skippedReason === 'not_found' || !moved.movedCard) {
    addToast('Unable to move card to stack');
    return;
  }

  const updatedHand = moved.hand;
  const updatedStack = moved.stack;
  const updatedPlayers = g.Players.map(p => {
    if (p.Username !== user) return p;
    return { ...p, Boardstate: { ...p.Boardstate, Hand: updatedHand } };
  });

  const input = buildGameInputFromGame(g, updatedPlayers, updatedStack);

  try {
    await apolloClient.mutate({
      mutation: UPDATE_GAME_MUTATION,
      variables: { input },
    });
    games.activeGame = { ...g, Players: updatedPlayers, Stack: updatedStack } as any;
    addToast(`Put ${moved.movedCard.Name} on stack`);
  } catch (e) {
    console.error('[board] moveHandCardToStack failed', e);
    addToast('Failed to put card on stack');
  }
}

async function resolveStackCard(stackIndex: number) {
  const g = game.value;
  if (!g) return;
  const stack = g.Stack ?? [];
  const card = stack[stackIndex];
  if (!card) return;
  const owner = card.CurrentZone;
  if (!owner) {
    addToast('Unknown card owner');
    return;
  }
  const ownerPlayer = g.Players.find(p => p.Username === owner);
  if (!ownerPlayer?.Boardstate) {
    addToast('Owner not found');
    return;
  }
  const graveyard = ownerPlayer.Boardstate.Graveyard ?? [];
  const resolved = resolveStackCardToGraveyardState(stack, graveyard, stackIndex);
  if (resolved.skippedReason || !resolved.movedCard) {
    addToast('Unable to resolve stack card');
    return;
  }

  const updatedPlayers = g.Players.map(p => {
    if (p.Username !== owner) return p;
    return { ...p, Boardstate: { ...p.Boardstate, Graveyard: resolved.graveyard } };
  });
  const input = buildGameInputFromGame(g, updatedPlayers, resolved.stack);

  try {
    await apolloClient.mutate({
      mutation: UPDATE_GAME_MUTATION,
      variables: { input },
    });
    games.activeGame = { ...g, Players: updatedPlayers, Stack: resolved.stack } as any;
    addToast(`Resolved ${resolved.movedCard.Name}`);
  } catch (e) {
    console.error('[board] resolveStackCard failed', e);
    addToast('Failed to resolve stack card');
  }
}

// Click vs dblclick handling to avoid triggering quick-move on double-tap
const clickTimers = new Map<string, number>();
function onCardClick(card: any, user: string, zone: Zone, fromIndex?: number) {
  const key = card.ID || `${user}:${zone}:${fromIndex ?? card.Name}`;
  if (clickTimers.has(key)) {
    window.clearTimeout(clickTimers.get(key)!);
    clickTimers.delete(key);
  }
  const t = window.setTimeout(() => {
    quickMove(card, user, zone, fromIndex);
    clickTimers.delete(key);
  }, 200);
  clickTimers.set(key, t as unknown as number);
}
function onCardDblClick(card: any, user: string, zone: Zone, fromIndex?: number) {
  const key = card.ID || `${user}:${zone}:${fromIndex ?? card.Name}`;
  if (clickTimers.has(key)) {
    window.clearTimeout(clickTimers.get(key)!);
    clickTimers.delete(key);
  }
  if (zone === 'Hand') {
    if (!isSelf(user)) {
      addToast('Only your hand can be played');
      return;
    }
    if (!hasPriority.value) {
      addToast(`Only ${currentPriority.value ?? 'the priority player'} can add to the stack`);
      return;
    }
    if (isLandCard(card)) {
      addToast("Lands can't be put on the stack");
      return;
    }
    moveHandCardToStack(card, user, fromIndex);
    return;
  }
  if (zone !== 'Battlefield') {
    addToast('Only battlefield cards can be tapped');
    return;
  }
  toggleTapped(user, zone, card, fromIndex);
}

// Toolbar actions (self only)
async function draw(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  await moveCard({
    gameID: g.ID,
    user: username,
    fromUser: username,
    cardID: top.ID,
    fromZone: 'Library',
    toZone: 'Hand',
    fromIndex: 0,
  });
  addToast(`Drew ${top.Name}`);
}

async function mill(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  await moveCard({
    gameID: g.ID,
    user: username,
    fromUser: username,
    cardID: top.ID,
    fromZone: 'Library',
    toZone: 'Graveyard',
    fromIndex: 0,
  });
  addToast(`Milled ${top.Name}`);
}

async function revealTop(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  await moveCard({
    gameID: g.ID,
    user: username,
    fromUser: username,
    cardID: top.ID,
    fromZone: 'Library',
    toZone: 'Revealed',
    fromIndex: 0,
  });
  addToast(`Revealed ${top.Name}`);
}

async function shuffleLibrary(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  if (!player?.Boardstate?.Library || player.Boardstate.Library.length < 2) return;

  // Fisher-Yates shuffle (new array)
  const shuffled = [...player.Boardstate.Library];
  for (let i = shuffled.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [shuffled[i], shuffled[j]] = [shuffled[j], shuffled[i]];
  }

  const input: any = {
    ...buildBoardstateInput(player, username, g.ID, { Library: shuffled }, player.Boardstate.Life),
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    addToast('Shuffled library');
    // Optimistic local patch
    applyLocalBoardstatePatch(username, (prev) => ({
      ...prev,
      ...input,
    }));
  } catch (e) {
    console.error('[board] shuffleLibrary failed', e);
  }
}

async function changeLife(username: string, delta: number) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  if (!player?.Boardstate) return;

  const input: any = {
    ...buildBoardstateInput(player, username, g.ID, {}, (player.Boardstate.Life ?? 0) + delta),
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    addToast(`${delta > 0 ? 'Gained' : 'Lost'} 1 life`);
    applyLocalBoardstatePatch(username, (prev) => ({ ...prev, ...input }));
  } catch (e) {
    console.error('[board] changeLife failed', e);
  }
}

// Scry 1 UX
const scry = ref<{ open: boolean; username: string; card?: { ID: string; Name: string } } | null>(null);

function scryOne(username: string) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === username);
  const top = player?.Boardstate?.Library?.[0];
  if (!top) return;
  scry.value = { open: true, username, card: { ID: top.ID, Name: top.Name } };
}

function scryKeepTop() {
  if (!scry.value) return;
  addToast(`Kept ${scry.value.card?.Name} on top`);
  scry.value = null;
}

async function scryPutBottom() {
  const g = game.value;
  const s = scry.value;
  if (!g || !s) return;
  const player = g.Players.find(p => p.Username === s.username);
  if (!player?.Boardstate?.Library || player.Boardstate.Library.length === 0) return;

  const [, ...rest] = player.Boardstate.Library;
  const newLibrary = [...rest, player.Boardstate.Library[0]];

  const input: any = {
    ...buildBoardstateInput(player, s.username, g.ID, { Library: newLibrary }, player.Boardstate.Life),
  };

  try {
    await apolloClient.mutate({
      mutation: UPDATE_BOARDSTATE_MUTATION,
      variables: { input },
    });
    addToast(`Put ${s.card?.Name} on bottom`);
    applyLocalBoardstatePatch(s.username, (prev) => ({ ...prev, ...input }));
  } catch (e) {
    console.error('[board] scryPutBottom failed', e);
  } finally {
    scry.value = null;
  }
}

// Helper: patch a player's boardstate in the local active game
function applyLocalBoardstatePatch(username: string, updater: (prev: any) => any) {
  const root = games.activeGame as any;
  if (!root) return;
  const updatedPlayers = root.Players.map((p: any) => {
    if (p.Username !== username) return p;
    const prev = p.Boardstate ? { ...p.Boardstate } : {};
    const next = updater(prev);
    return { ...p, Boardstate: next };
  });
  // Replace the entire activeGame object to avoid mutating frozen Apollo results
  games.activeGame = { ...root, Players: updatedPlayers } as any;
}

// Toasts
type Toast = { id: number; text: string };
const toasts = ref<Toast[]>([]);
let toastCounter = 0;
function addToast(text: string, duration = toastDurationMs.value) {
  const id = ++toastCounter;
  toasts.value.push({ id, text });
  window.setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id);
  }, duration);
}

// Per-player per-zone view toggle (tiles vs art)
const stackedZones = ref<Record<string, Record<string, boolean>>>({});
function toggleStack(user: string, zone: string) {
  if (!stackedZones.value[user]) stackedZones.value[user] = {};
  stackedZones.value[user][zone] = !stackedZones.value[user][zone];
}
function isStacked(user: string, zone: string) {
  return !!(stackedZones.value[user] && stackedZones.value[user][zone]);
}

// Persist view preferences in localStorage (per user+zone)
const STACKED_ZONES_KEY = 'vedh:stackedZones:v1';
onMounted(() => {
  try {
    const raw = localStorage.getItem(STACKED_ZONES_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      if (parsed && typeof parsed === 'object') {
        stackedZones.value = parsed;
      }
    }
  } catch {}
});
watch(stackedZones, (val) => {
  try {
    localStorage.setItem(STACKED_ZONES_KEY, JSON.stringify(val));
  } catch {}
}, { deep: true });

</script>

<style scoped lang="scss">
.board {
  /* Make the board take the full viewport so we can anchor the main player */
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 100vh;
  --main-player-height: 33vh; /* bottom third reserved for player's control center */
  --turn-accent: #f5b342;
}

.board-header {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 1rem 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.board-header .title {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.turn-spotlight {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.02));
  border: 1px solid rgba(255, 255, 255, 0.22);
  box-shadow: 0 14px 28px rgba(0, 0, 0, 0.32), inset 0 0 0 1px rgba(255, 255, 255, 0.05);
  transition: box-shadow 0.25s ease, border-color 0.25s ease, background 0.25s ease;
}

.turn-spotlight .eyebrow {
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  opacity: 0.7;
}

.turn-spotlight strong {
  display: block;
  font-size: 1.35rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.turn-number,
.turn-phase,
.turn-priority {
  display: inline-flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 6rem;
}

.turn-priority strong {
  color: var(--turn-accent);
}

.turn-spotlight.priority-owner {
  border-color: rgba(245, 179, 66, 0.6);
  background: linear-gradient(135deg, rgba(245, 179, 66, 0.22), rgba(255, 255, 255, 0.04));
  box-shadow:
    0 14px 28px rgba(0, 0, 0, 0.35),
    0 0 18px rgba(245, 179, 66, 0.35),
    0 0 36px rgba(245, 179, 66, 0.2);
}

.turn-meta {
  margin: 0;
  color: rgba(255, 255, 255, 0.78);
  font-size: 0.85rem;
}

.turn-tracker {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: 0.5rem;
}

.turn-tracker .chip {
  font-size: 0.75rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 0.25rem 0.5rem;
  border-radius: 999px;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.1);
  color: rgba(255,255,255,0.8);
}

.turn-controls {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
}

.win-claim {
  display: inline-flex;
  gap: 0.6rem;
  align-items: center;
  margin-top: 0.35rem;
  font-size: 0.75rem;
  color: rgba(255,255,255,0.8);
}

.win-claim .muted {
  color: rgba(255,255,255,0.6);
}

.settings {
  position: relative;
}

.settings-trigger {
  appearance: none;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
  font-size: 1rem;
  padding: 0.35rem 0.6rem;
  border-radius: 10px;
  cursor: pointer;
}

.settings-drawer {
  position: absolute;
  right: 0;
  top: calc(100% + 0.5rem);
  background: rgba(18, 18, 18, 0.98);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 12px;
  min-width: 220px;
  padding: 0.75rem 0.9rem;
  display: grid;
  gap: 0.6rem;
  z-index: 12;
}

.settings-title {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(255,255,255,0.6);
}

.toast-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.85rem;
  color: rgba(255,255,255,0.8);
}

.toast-toggle input {
  accent-color: rgba(133, 215, 255, 0.9);
}

.toast-ttl {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: rgba(255,255,255,0.75);
}

.toast-ttl input {
  width: 90px;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 8px;
  color: #fff;
  padding: 0.2rem 0.4rem;
}

.toast-ttl .unit {
  opacity: 0.7;
}

.menu { position: relative; }
.menu-trigger {
  appearance: none;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
  font-size: 0.85rem;
  padding: 0.35rem 0.7rem;
  border-radius: 8px;
}
.menu-popover {
  position: absolute;
  right: 0;
  margin-top: 0.4rem;
  background: rgba(30,30,30,0.98);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 10px;
  min-width: 220px;
  padding: 0.6rem 0.75rem;
  z-index: 10;
}
.menu-section { padding: 0.4rem 0; }
.menu-title { font-size: 0.75rem; opacity: 0.8; text-transform: uppercase; letter-spacing: 0.06em; margin-bottom: 0.25rem; }
.menu-section label { display: flex; align-items: center; gap: 0.4rem; padding: 0.2rem 0; font-size: 0.9rem; }

.board-grid {
  /* Scrollable area that contains opponents/stack. Reserve space for the anchored main-player. */
  flex: 1 1 auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  overflow: auto;
  /* Reserve room at the bottom so content doesn't get hidden behind the anchored main-player */
  padding-bottom: calc(var(--main-player-height) + 1rem);
}

.players {
  display: grid;
  gap: 0.75rem;
}

.players article {
  display: grid;
  grid-template-columns: repeat(6, minmax(140px, 1fr));
  gap: 0.75rem;
  align-items: flex-start;
  overflow-x: auto;
}

.players article {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 14px;
  padding: 0.75rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.players article.active {
  border-color: rgba(133, 215, 255, 0.6);
  box-shadow: 0 0 0 1px rgba(133, 215, 255, 0.15);
}

.players article > header {
  grid-column: 1 / -1;
}

.players article .zone {
  min-width: 140px;
}

.players article .zone[data-zone='Battlefield'] {
  grid-column: 1 / -1;
}

.players article .zone[data-zone='Hand'] {
  grid-column: 1 / -1;
}

.players article .zone[data-zone='Commander'] { grid-column: 1; }
.players article .zone[data-zone='Graveyard'] { grid-column: 2; }
.players article .zone[data-zone='Exiled'] { grid-column: 3; }
.players article .zone[data-zone='Revealed'] { grid-column: 4; }
.players article .zone[data-zone='Controlled'] { grid-column: 5; }
.players article .zone[data-zone='Library'] { grid-column: 6; }

.zone {
  margin-top: 0.5rem;
  border: 1px solid rgba(255,255,255,0.04);
  background: rgba(0,0,0,0.02);
  padding: 0.5rem;
  border-radius: 8px;
  position: relative;
  transition: box-shadow 140ms ease, border-color 120ms ease;
}

/* drop indicator and stronger drag-over visuals */
.zone::before {
  content: '';
  position: absolute;
  left: 10%;
  right: 10%;
  top: 8px;
  height: 4px;
  border-radius: 4px;
  background: transparent;
  opacity: 0;
  transition: background 160ms ease, opacity 160ms ease, transform 160ms ease;
}
.zone.drag-over::before {
  background: linear-gradient(90deg, rgba(133,215,255,0.95), rgba(80,180,255,0.85));
  opacity: 1;
  transform: scaleX(1);
}
.zone.drag-over {
  border-color: rgba(80,180,255,0.9);
  box-shadow: 0 12px 36px rgba(6,20,30,0.6);
}

.zone h3 {
  margin: 0 0 0.25rem;
  font-size: 0.8rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgba(255, 255, 255, 0.65);
}

.zone h3 small {
  text-transform: none;
  letter-spacing: normal;
  font-weight: normal;
  font-size: 0.8em;
  color: rgba(255, 255, 255, 0.5);
}

.zone ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: grid;
  gap: 0.25rem;
}

/* Tiles view */
.cards.tiles {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 0.5rem;
}

/* Art view (image-only) */
.cards.art {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 0.6rem;
}

/* Prioritize battlefield + hand as horizontal rows */
.zone[data-zone='Battlefield'] .cards.tiles,
.zone[data-zone='Hand'] .cards.tiles,
.zone[data-zone='Battlefield'] .cards.art,
.zone[data-zone='Hand'] .cards.art,
.zone[data-zone='Hand'] .hand {
  display: flex;
  flex-wrap: nowrap;
  align-items: stretch;
  gap: 0.6rem;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 0.25rem;
}

.zone[data-zone='Battlefield'] .cards.tiles > *,
.zone[data-zone='Hand'] .cards.tiles > *,
.zone[data-zone='Battlefield'] .cards.art > *,
.zone[data-zone='Hand'] .cards.art > * {
  flex: 0 0 120px;
}

.zone[data-zone='Battlefield'] .cards.art > *,
.zone[data-zone='Hand'] .cards.art > * {
  flex-basis: 140px;
}

/* Keep secondary zones in a row with vertical card stacks */
.zone:not([data-zone='Battlefield']):not([data-zone='Hand']) .cards.tiles,
.zone:not([data-zone='Battlefield']):not([data-zone='Hand']) .cards.art {
  grid-template-columns: 1fr;
}

/* Allow zone lists to expand vertically to fit their cards */
.zone ul {
  display: grid; /* keep grid behavior for tiles but let it grow */
  grid-auto-rows: auto;
}
.card-tile {
  display: grid;
  gap: 0.35rem;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.4rem;
  cursor: grab;
}
.card-tile img {
  width: 100%;
  aspect-ratio: 0.714; /* 63x88mm ratio */
  object-fit: cover;
  border-radius: 6px;
}
.card-tile .label { font-size: 0.8rem; opacity: 0.9; }

.stack-card {
  position: relative;
}
.stack-resolve {
  position: absolute;
  top: 6px;
  right: 6px;
  z-index: 2;
  border: 1px solid rgba(255, 255, 255, 0.2);
  background: rgba(20, 20, 20, 0.85);
  color: #fff;
  font-size: 0.7rem;
  padding: 0.2rem 0.4rem;
  border-radius: 6px;
  cursor: pointer;
}
.stack-resolve:hover {
  background: rgba(40, 40, 40, 0.9);
}

/* Pulsing glow when dragging or on hover */
@keyframes pulse-glow {
  0% { box-shadow: 0 6px 18px rgba(133,215,255,0.0); }
  50% { box-shadow: 0 10px 30px rgba(133,215,255,0.25); }
  100% { box-shadow: 0 6px 18px rgba(133,215,255,0.0); }
}
.card-tile { position: relative; transition: transform 140ms ease, box-shadow 160ms ease; }
.card-tile.dragging {
  transform: translateY(-6px) scale(1.02);
  animation: pulse-glow 1.2s ease-in-out infinite;
  z-index: 200;
}
.card-tile:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 22px rgba(0,0,0,0.55);
}

/* Stacks view */
.cards.stacks {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 0.5rem;
}
.stack-group {
  position: relative;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 0.5rem;
  cursor: grab;
}
.stack-condensed {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}
.stack-thumb {
  position: relative;
  width: 56px;
  height: 80px;
  flex: 0 0 56px;
  --gap: 8px; /* collapsed gap between cards */
  --hover-gap: 16px; /* expanded gap on hover */
}
.stack-thumb img {
  position: absolute;
  left: 0;
  width: 56px;
  height: 80px;
  object-fit: cover;
  border-radius: 6px;
  box-shadow: 0 2px 6px rgba(0,0,0,0.35);
  transition: top 160ms ease, transform 140ms ease;
  top: calc(var(--i) * var(--gap));
}
.stack-group:hover .stack-thumb { --gap: var(--hover-gap); }
.stack-info { flex: 1 1 auto; min-width: 0; }
.stack-name { font-size: 0.9rem; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.stack-meta { font-size: 0.75rem; opacity: 0.7; margin-top: 0.15rem; }
.stack-group .count {
  background: rgba(0,0,0,0.6);
  color: #fff;
  padding: 0.12rem 0.5rem;
  border-radius: 999px;
  font-size: 0.75rem;
  margin-left: 0.5rem;
}

/* Battlefield type groups */
.bf-group { margin-bottom: 0.75rem; }
.bf-group-title { font-size: 0.8rem; opacity: 0.8; margin: 0.15rem 0 0.35rem; text-transform: uppercase; letter-spacing: 0.06em; }

.card {
  padding: 0.2rem 0.45rem;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  font-size: 0.9rem;
  cursor: grab;
  user-select: none;
}

.card:active {
  cursor: grabbing;
}

.muted {
  opacity: 0.7;
}

.stack {
  background: rgba(255, 255, 255, 0.04);
  border-radius: 14px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0.75rem 1rem;
  display: grid;
  gap: 0.5rem;
  position: sticky;
  top: 0.75rem;
  z-index: 5;
}

@keyframes stack-pop {
  0% { transform: translateY(0); box-shadow: 0 0 0 rgba(133,215,255,0); border-color: rgba(255,255,255,0.08); }
  20% { transform: translateY(-4px); box-shadow: 0 12px 28px rgba(133,215,255,0.45); border-color: rgba(133,215,255,0.8); }
  55% { transform: translateY(0); box-shadow: 0 10px 26px rgba(80,180,255,0.3); border-color: rgba(133,215,255,0.45); }
  100% { transform: translateY(0); box-shadow: 0 0 0 rgba(133,215,255,0); border-color: rgba(255,255,255,0.08); }
}

.stack.pulse {
  animation: stack-pop 1600ms ease;
}

/* Anchor main player to the bottom of the viewport (within the scroll container) */
.main-player {
  /* Fixed to viewport bottom so it's always visible and takes the lower third */
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 999;
  height: var(--main-player-height);
  background: linear-gradient(180deg, rgba(24,24,24,0.98) 0%, rgba(12,12,12,0.98) 100%);
  border-top: 4px solid rgba(255,255,255,0.06); /* sharp dividing line */
  padding: 0;
  border-radius: 0 0 0 0;
  box-shadow: 0 -14px 40px rgba(0,0,0,0.55);
  overflow: hidden; /* keep resize handle pinned while content scrolls beneath it */
}

/* Drag handle / sharper divider */
.main-player-resize-handle {
  position: absolute;
  left: 0;
  right: 0;
  top: 0;
  height: 10px;
  cursor: ns-resize;
  background: linear-gradient(90deg, rgba(255,255,255,0.08), rgba(255,255,255,0.02));
}

/* Constrain inner content so it lines up with the rest of the app while the background spans full width */
.main-player > article {
  max-width: 1200px;
  margin: 10px auto 0;
  padding: 0.5rem 1rem 0;
  display: grid;
  grid-template-columns: minmax(220px, 220px) 1fr;
  grid-auto-rows: min-content;
  gap: 0.75rem;
  align-items: start;
  height: calc(100% - 10px);
  box-sizing: border-box;
  overflow: auto;
  -webkit-overflow-scrolling: touch;
}

.main-player-left {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.main-player-header {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.main-player-header .life-row {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.main-player-right {
  display: grid;
  grid-template-columns: repeat(6, minmax(130px, 1fr));
  gap: 0.75rem;
  align-items: start;
  overflow-x: auto;
}

.main-player-right .zone[data-zone='Battlefield'] {
  grid-column: 1 / -1;
}

.main-player-right .zone[data-zone='Hand'] {
  grid-column: 1 / -1;
}

.main-player-right .zone[data-zone='Graveyard'] { grid-column: 1; }
.main-player-right .zone[data-zone='Exiled'] { grid-column: 2; }
.main-player-right .zone[data-zone='Revealed'] { grid-column: 3; }
.main-player-right .zone[data-zone='Controlled'] { grid-column: 4; }
.main-player-right .zone[data-zone='Library'] { grid-column: 5; }

/* Footer zones should behave like opponent zones */
.main-player .zone {
  max-height: calc(var(--main-player-height) - 64px);
  overflow: auto; /* allow zone-level scrolling when content overflows vertically */
}

.main-player .zone[data-zone='Battlefield'],
.main-player .zone[data-zone='Hand'] {
  max-height: none;
}

.main-player-left .zone {
  max-height: none;
}

.main-player .cards.tiles {
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
}

.main-player .cards.art {
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
}

@media (max-width: 900px) {
  .main-player > article {
    grid-template-columns: 1fr;
  }
}

.main-player-collapsed-toggle {
  position: fixed;
  left: 50%;
  bottom: 0.6rem;
  transform: translateX(-50%);
  z-index: 999;
  appearance: none;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(20, 20, 20, 0.9);
  color: #fff;
  font-size: 0.9rem;
  padding: 0.25rem 0.6rem;
  border-radius: 999px;
  box-shadow: 0 -6px 18px rgba(0,0,0,0.4);
  cursor: pointer;
}

.stack ol {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 0.9rem;
}

.loading-state {
  display: grid;
  place-items: center;
  min-height: 60vh;
  color: rgba(255, 255, 255, 0.6);
}

/* Toolbar */
header .player-toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  margin-top: 0.4rem;
}

.player-toolbar .tool {
  appearance: none;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.06);
  color: #fff;
  font-size: 0.8rem;
  padding: 0.25rem 0.5rem;
  border-radius: 999px;
}

.player-toolbar .life-tools {
  display: inline-flex;
  gap: 0.3rem;
  margin-left: auto;
}

.life-tools.inline {
  margin-left: 0;
}

.player-toolbar .tool:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Scry modal */
.scry-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: grid;
  place-items: center;
}

.scry-modal {
  background: rgba(30, 30, 30, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 1rem 1.25rem;
  min-width: 260px;
  max-width: 420px;
}

.scry-modal header {
  font-weight: 600;
  margin-bottom: 0.5rem;
}

.scry-card {
  display: block;
  width: min(220px, 70vw);
  aspect-ratio: 200 / 280;
  object-fit: cover;
  border-radius: 8px;
  margin: 0.25rem auto 0.5rem;
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.45);
}

.scry-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.75rem;
}

/* Toasts */
.toasts {
  position: fixed;
  right: 1rem;
  bottom: calc(var(--main-player-height) + 1rem);
  display: grid;
  gap: 0.5rem;
  z-index: 10;
}

.toasts.compact {
  bottom: 1rem;
}

.toast {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #fff;
  padding: 0.5rem 0.75rem;
  border-radius: 8px;
  font-size: 0.9rem;
}
</style>
