<template>
  <section class="board" v-if="game" :style="{ '--main-player-height': mainPlayerHeightCss }">
    <header class="board-header">
      <div class="title">
        <h1>Game {{ game.ID }}</h1>
        <p>Turn {{ game.Turn?.Number ?? '—' }} • {{ game.Turn?.Player ?? 'Unknown player' }} ({{ game.Turn?.Phase ?? 'Phase' }})</p>
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
                  {{ isStacked(player.Username, 'Commander') ? 'Tiles' : 'Stack' }}
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
                <ul class="cards stacks">
                  <li v-for="g in groupByName(player.Boardstate?.Commander ?? [])" :key="g.name" class="stack-group">
                    <div class="stack-condensed">
                      <div class="stack-thumb">
                        <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                      </div>
                      <div class="stack-info">
                        <div class="stack-name">{{ g.name }}</div>
                        <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                      </div>
                      <div class="count">{{ g.count }}</div>
                    </div>
                  </li>
                </ul>
              </template>
          </div>
          <div class="zone" :data-zone="'Battlefield'" :class="{ 'drag-over': isDragOver(player.Username, 'Battlefield') }" @dragenter.prevent="onDragEnter(player.Username, 'Battlefield')" @dragleave.prevent="onDragLeave(player.Username, 'Battlefield')">
              <h3>
                Battlefield
                <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Battlefield')">
                  {{ isStacked(player.Username, 'Battlefield') ? 'Tiles' : 'Stack' }}
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
                <ul class="cards stacks">
                  <li v-for="g in groupByName(player.Boardstate?.Battlefield ?? [])" :key="g.name" class="stack-group">
                    <div class="stack-condensed">
                      <div class="stack-thumb">
                        <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                      </div>
                      <div class="stack-info">
                        <div class="stack-name">{{ g.name }}</div>
                        <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                      </div>
                      <div class="count">{{ g.count }}</div>
                    </div>
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
                {{ isStacked(player.Username, 'Graveyard') ? 'Tiles' : 'Stack' }}
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
              <ul class="cards stacks">
                <li v-for="g in groupByName(player.Boardstate?.Graveyard ?? [])" :key="g.name" class="stack-group">
                  <div class="stack-condensed">
                    <div class="stack-thumb">
                      <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                    </div>
                    <div class="stack-info">
                      <div class="stack-name">{{ g.name }}</div>
                      <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                    </div>
                    <div class="count">{{ g.count }}</div>
                  </div>
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Exiled'" :class="{ 'drag-over': isDragOver(player.Username, 'Exiled') }" @dragenter.prevent="onDragEnter(player.Username, 'Exiled')" @dragleave.prevent="onDragLeave(player.Username, 'Exiled')">
            <h3>
              Exiled ({{ player.Boardstate?.Exiled?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Exiled')">
                {{ isStacked(player.Username, 'Exiled') ? 'Tiles' : 'Stack' }}
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
              <ul class="cards stacks">
                <li v-for="g in groupByName(player.Boardstate?.Exiled ?? [])" :key="g.name" class="stack-group">
                  <div class="stack-condensed">
                    <div class="stack-thumb">
                      <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                    </div>
                    <div class="stack-info">
                      <div class="stack-name">{{ g.name }}</div>
                      <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                    </div>
                    <div class="count">{{ g.count }}</div>
                  </div>
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Revealed'" :class="{ 'drag-over': isDragOver(player.Username, 'Revealed') }" @dragenter.prevent="onDragEnter(player.Username, 'Revealed')" @dragleave.prevent="onDragLeave(player.Username, 'Revealed')">
            <h3>
              Revealed ({{ player.Boardstate?.Revealed?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Revealed')">
                {{ isStacked(player.Username, 'Revealed') ? 'Tiles' : 'Stack' }}
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
              <ul class="cards stacks">
                <li v-for="g in groupByName(player.Boardstate?.Revealed ?? [])" :key="g.name" class="stack-group">
                  <div class="stack-condensed">
                    <div class="stack-thumb">
                      <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                    </div>
                    <div class="stack-info">
                      <div class="stack-name">{{ g.name }}</div>
                      <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                    </div>
                    <div class="count">{{ g.count }}</div>
                  </div>
                </li>
              </ul>
            </template>
          </div>
          <div class="zone" :data-zone="'Controlled'" :class="{ 'drag-over': isDragOver(player.Username, 'Controlled') }" @dragenter.prevent="onDragEnter(player.Username, 'Controlled')" @dragleave.prevent="onDragLeave(player.Username, 'Controlled')">
            <h3>
              Controlled ({{ player.Boardstate?.Controlled?.length ?? 0 }})
              <button class="tool" style="margin-left:0.5rem; font-size:0.7rem; padding:0.15rem 0.4rem;" @click="toggleStack(player.Username, 'Controlled')">
                {{ isStacked(player.Username, 'Controlled') ? 'Tiles' : 'Stack' }}
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
              <ul class="cards stacks">
                <li v-for="g in groupByName(player.Boardstate?.Controlled ?? [])" :key="g.name" class="stack-group">
                  <div class="stack-condensed">
                    <div class="stack-thumb">
                      <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                    </div>
                    <div class="stack-info">
                      <div class="stack-name">{{ g.name }}</div>
                      <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                    </div>
                    <div class="count">{{ g.count }}</div>
                  </div>
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
      <section class="stack">
        <header>
          <h2>Stack</h2>
        </header>
        <ul class="cards tiles">
          <Card
            v-for="(card, i) in game.Stack"
            :key="`${card.ID}-${i}`"
            :id="card.ID"
            :name="card.Name"
            :image-src="getImage(card.Name)"
            :draggable="false"
          />
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
              <ul class="cards stacks">
                <li v-for="g in groupByName(selfPlayer.Boardstate?.Commander ?? [])" :key="g.name" class="stack-group" draggable="true" @dragstart="onStackDragStart(selfPlayer.Username, 'Commander', g)" @dragend="onDragEnd">
                  <div class="stack-condensed">
                    <div class="stack-thumb">
                        <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                    </div>
                    <div class="stack-info">
                      <div class="stack-name">{{ g.name }}</div>
                      <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                    </div>
                    <div class="count">{{ g.count }}</div>
                  </div>
                </li>
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
            <ul class="cards stacks">
              <li v-for="g in groupByName(selfPlayer.Boardstate?.Battlefield ?? [])" :key="g.name" class="stack-group" draggable="true" @dragstart="onStackDragStart(selfPlayer.Username, 'Battlefield', g)" @dragend="onDragEnd">
                <div class="stack-condensed">
                  <div class="stack-thumb">
                    <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                  </div>
                  <div class="stack-info">
                    <div class="stack-name">{{ g.name }}</div>
                    <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                  </div>
                  <div class="count">{{ g.count }}</div>
                </div>
              </li>
            </ul>
          </template>
        </div>
  <div class="zone" :data-zone="'Hand'" :class="{ 'drag-over': isDragOver(selfPlayer.Username, 'Hand') }" @dragenter.prevent="onDragEnter(selfPlayer.Username, 'Hand')" @dragleave.prevent="onDragLeave(selfPlayer.Username, 'Hand')" @dragover.prevent @drop.prevent="onDrop(selfPlayer.Username, 'Hand')">
          <h3>Hand ({{ selfPlayer.Boardstate?.Hand?.length ?? 0 }})</h3>
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
            <ul class="cards stacks">
              <li v-for="g in groupByName(selfPlayer.Boardstate?.Graveyard ?? [])" :key="g.name" class="stack-group" draggable="true" @dragstart="onStackDragStart(selfPlayer.Username, 'Graveyard', g)" @dragend="onDragEnd">
                <div class="stack-condensed">
                  <div class="stack-thumb">
                    <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                  </div>
                  <div class="stack-info">
                    <div class="stack-name">{{ g.name }}</div>
                    <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                  </div>
                  <div class="count">{{ g.count }}</div>
                </div>
              </li>
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
            <ul class="cards stacks">
              <li v-for="g in groupByName(selfPlayer.Boardstate?.Exiled ?? [])" :key="g.name" class="stack-group" draggable="true" @dragstart="onStackDragStart(selfPlayer.Username, 'Exiled', g)" @dragend="onDragEnd">
                <div class="stack-condensed">
                  <div class="stack-thumb">
                    <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                  </div>
                  <div class="stack-info">
                    <div class="stack-name">{{ g.name }}</div>
                    <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                  </div>
                  <div class="count">{{ g.count }}</div>
                </div>
              </li>
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
            <ul class="cards stacks">
              <li v-for="g in groupByName(selfPlayer.Boardstate?.Revealed ?? [])" :key="g.name" class="stack-group" draggable="true" @dragstart="onStackDragStart(selfPlayer.Username, 'Revealed', g)" @dragend="onDragEnd">
                <div class="stack-condensed">
                  <div class="stack-thumb">
                    <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                  </div>
                  <div class="stack-info">
                    <div class="stack-name">{{ g.name }}</div>
                    <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                  </div>
                  <div class="count">{{ g.count }}</div>
                </div>
              </li>
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
            <ul class="cards stacks">
              <li v-for="g in groupByName(selfPlayer.Boardstate?.Controlled ?? [])" :key="g.name" class="stack-group" draggable="true" @dragstart="onStackDragStart(selfPlayer.Username, 'Controlled', g)" @dragend="onDragEnd">
                <div class="stack-condensed">
                  <div class="stack-thumb">
                    <img v-for="(c, idx) in g.sample.slice(0,4)" :key="c.ID || idx" :src="getImage(c.Name)" :alt="c.Name" :style="{ '--i': idx, zIndex: 10 - idx }" />
                  </div>
                  <div class="stack-info">
                    <div class="stack-name">{{ g.name }}</div>
                    <div class="stack-meta">{{ g.sample[0]?.Types ?? '' }}</div>
                  </div>
                  <div class="count">{{ g.count }}</div>
                </div>
              </li>
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
  <div class="toasts">
    <div class="toast" v-for="t in toasts" :key="t.id">{{ t.text }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useGamesStore } from '../stores/games';
import { useAuthStore } from '../stores/auth';
import { apolloClient } from '../services/apollo';
import { UPDATE_BOARDSTATE_MUTATION } from '../graphql/mutations';
// Subscriptions are handled centrally in the games store.
import { fetchScryfallImageByName } from '../services/scryfall';
import Card from '../components/Card.vue';
// Dev logging helper: use console.log so messages appear without enabling Verbose level
function dbg(...args: any[]) { console.log(...args); }

const games = useGamesStore();
const auth = useAuthStore();
const route = useRoute();

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

// Simple tile-only view; no display toggles needed
const MAIN_PLAYER_PREF_KEY = 'vedh:mainPlayerPanel:v1';
const mainPlayerHeight = ref(33);
const mainPlayerCollapsed = ref(false);
const mainPlayerHeightCss = computed(() => mainPlayerCollapsed.value ? '0px' : `${mainPlayerHeight.value}vh`);
const mainPlayerResizing = ref(false);
const mainPlayerResizeStart = ref({ y: 0, height: 33 });
const MAIN_PLAYER_MIN_VH = 16;
const MAIN_PLAYER_MAX_VH = 60;

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
});
watch([mainPlayerHeight, mainPlayerCollapsed], () => {
  try {
    localStorage.setItem(MAIN_PLAYER_PREF_KEY, JSON.stringify({
      height: mainPlayerHeight.value,
      collapsed: mainPlayerCollapsed.value,
    }));
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
    destList.push(nextCard);
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
function addToast(text: string, duration = 2500) {
  const id = ++toastCounter;
  toasts.value.push({ id, text });
  window.setTimeout(() => {
    toasts.value = toasts.value.filter(t => t.id !== id);
  }, duration);
}

// Per-player per-zone "stacked" mode: shows grouped names (counts) instead of full tiles
const stackedZones = ref<Record<string, Record<string, boolean>>>({});
function toggleStack(user: string, zone: string) {
  if (!stackedZones.value[user]) stackedZones.value[user] = {};
  stackedZones.value[user][zone] = !stackedZones.value[user][zone];
}
function isStacked(user: string, zone: string) {
  return !!(stackedZones.value[user] && stackedZones.value[user][zone]);
}

function groupByName(list: { ID?: string; Name: string }[] | undefined) {
  const out: Array<{ name: string; count: number; sample: any[] }> = [];
  if (!list || list.length === 0) return out;
  const map: Record<string, { count: number; sample: { ID?: string; Name: string }[] }> = {};
  for (const c of list) {
    if (!map[c.Name]) map[c.Name] = { count: 0, sample: [] };
    map[c.Name].count++;
    if (map[c.Name].sample.length < 4) map[c.Name].sample.push(c);
  }
  for (const name of Object.keys(map)) {
    out.push({ name, count: map[name].count, sample: map[name].sample });
  }
  // sort alphabetically to keep stable order
  out.sort((a, b) => a.name.localeCompare(b.name));
  return out;
}

// Persist stacked view preferences in localStorage (per user+zone)
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

// Drag from a stack group: pick one concrete card from the player's zone with this name
function onStackDragStart(user: string, zone: Zone, group: { name: string; sample: any[] }) {
  const g = game.value;
  if (!g) return;
  const player = g.Players.find(p => p.Username === user);
  const list: any[] = (player?.Boardstate as any)?.[zone] ?? [];
  const foundIndex = list.findIndex(c => c?.Name === group.name);
  if (foundIndex === -1) return;
  const found = list[foundIndex];
  onDragStart({ ID: found.ID, Name: found.Name }, user, zone, foundIndex);
}
</script>

<style scoped lang="scss">
.board {
  /* Make the board take the full viewport so we can anchor the main player */
  display: flex;
  flex-direction: column;
  gap: 1rem;
  height: 100vh;
  --main-player-height: 33vh; /* bottom third reserved for player's control center */
}

.board-header {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 1rem 1.25rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.board-header .title p { margin: 0.25rem 0 0; color: rgba(255,255,255,0.7); }

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

/* Make each opponent card area responsive: zones should wrap horizontally and vertically
   and expand based on contained cards. */
.players article {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: flex-start;
}

.players article .zone {
  /* Each zone can grow/shrink and will wrap to the next row when needed */
  flex: 1 1 220px;
  min-width: 160px;
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
  padding: 0.5rem 0; /* remove side padding so it spans edge-to-edge visually */
  border-radius: 0 0 0 0;
  box-shadow: 0 -14px 40px rgba(0,0,0,0.55);
  overflow: auto; /* allow scrolling inside the control center */
  -webkit-overflow-scrolling: touch;
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
  margin: 0 auto;
  padding: 0 1rem;
  display: grid;
  grid-template-columns: minmax(220px, 220px) 1fr;
  grid-auto-rows: min-content;
  gap: 0.75rem;
  align-items: start;
  height: 100%;
  box-sizing: border-box;
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
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0.75rem;
  align-items: start;
}

/* Footer zones should behave like opponent zones */
.main-player .zone {
  max-height: calc(var(--main-player-height) - 64px);
  overflow: auto; /* allow zone-level scrolling when content overflows vertically */
}

.main-player-left .zone {
  max-height: none;
}

.main-player .cards.tiles {
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
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
  bottom: 1rem;
  display: grid;
  gap: 0.5rem;
  z-index: 10;
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
