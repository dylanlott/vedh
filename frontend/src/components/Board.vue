<template>
  <div class="board">
    <section class="myself">
      <!-- ### OPPONENTS -->
      <section v-if="self && opps" id="Opponents" class="opponents dropzone outer-dropzone">
        OPPONENTS
        <div v-for="opp in opps">
          <Opponent v-bind="{
            boardstate: opp.Boardstate,
            username: opp.Username,
          }"></Opponent>
        </div>
      </section>

      <!-- ### STACK -->
      <b-collapse aria-id="contentIdForA11y2" class="panel" animation="slide" v-model="isStackOpen">
        <template #trigger>
          <div class="panel-heading" role="button" aria-controls="contentIdForA11y2" :aria-expanded="isStackOpen">
            <strong>Stack</strong>
          </div>
        </template>
        <div id="Stack" class="panel-block stack dropzone outer-dropzone">
          <div class="container">
            <DraggableCard class="item" v-for="card in game.Stack" :card="card" :user="user" :handlers="{
              tap: handlers.tap,
              cast: handlers.cast,
            }" :key="card.ID" />
          </div>
        </div>
      </b-collapse>

      <!-- ### BATTLEFIELD -->
      <section v-if="self" id="Battlefield" class="battlefield dropzone outer-dropzone">
        BATTLEFIELD
        <DraggableCard class="item" v-for="card in self.Boardstate.Battlefield" :card="card" :user="user" :key="card.ID"
          :handlers="{
            tap: handlers.tap,
            cast: handlers.cast,
          }" />
      </section>

      <!-- ### HAND -->
      <section id="Hand" v-if="self" class="container hand dropzone outer-dropzone">
        HAND
        <DraggableCard class="item" v-for="card in self.Boardstate.Hand" :card="card" :user="user" :key="card.ID"
          :handlers="handlers" />
      </section>

      <!-- ### GRAVEYARD -->
      <div class="shell">
        <div class="columns">
          <section id="Graveyard" v-if="self" class="dropzone outer-dropzone column">
            GRAVEYARD
            <DraggableCard class="item" v-for="card in self.Boardstate.Graveyard" :card="card" :user="user"
              :key="card.ID" />
          </section>

          <!-- ### GRAVEYARD -->
          <section id="Exiled" v-if="self" class="dropzone outer-dropzone column">
            EXILED
            <DraggableCard class="item" v-for="card in self.Boardstate.Exiled" :card="card" :user="user" :key="card.ID" />
          </section>

          <section id="Revealed" v-if="self" class="dropzone outer-dropzone column">
            REVEALED
            <DraggableCard class="item" v-for="card in self.Boardstate.Revealed" :card="card" :user="user"
              :key="card.ID" />
          </section>
        </div>
      </div>
      <!-- INVITE LINK -->
      <section>
        <b-modal :active="isInviteModalOpen">
          <div v-if="self" class="modal-card" width="400px">
            <header class="modal-card-head">Get an Invite Link</header>
            <section class="modal-card-body">
              <code id="invite-link">{{ inviteLink() }}</code>
              <b-button type="is-primary" @click="copyToClipboard(inviteLink())">Copy</b-button>
            </section>
          </div>
        </b-modal>
      </section>
      <!-- END INVITE LINK -->

      <!-- COMMANDER SELECTION  -->
      <section>
        <b-modal has-modal-card :active="isCommanderSelectionOpen" trap-focus :destroy-on-hide="false" aria-role="dialog"
          aria-label="Commander Selection Modal" close-button-aria-label="Close" aria-modal>
          <div class="card">
            <div class="card-content">
              <div class="content" v-if="self">
                <b-field label="Select your Commander(s) from your library:">
                  <b-autocomplete v-model="searchQuery" placeholder="e.g. Kykar, Wind's Fury" :open-on-focus="openOnFocus"
                    :data="filteredCommanderData" field="Name" :clearable="clearable" @select="option => {
                      if (!!searchQuery) {
                        selectedCommander = option
                        addCommander(option)
                      }
                    }">
                  </b-autocomplete>
                </b-field>
                <b-field>
                  <b-tag :key="key" v-if="self.Boardstate.Commander.length > 0"
                    v-for="(obj, key) in self.Boardstate.Commander" type="is-primary" closable
                    aria-close-label="Close tag" @close="removeCommander(obj)">
                    <b>{{ obj.Name }}</b>
                  </b-tag>
                </b-field>
              </div>
            </div>
            <footer class="modal-card-foot">
              <b-button type="is-primary is-light" size="is-small"
                @click="isCommanderSelectionOpen = !isCommanderSelectionOpen">Done</b-button>
            </footer>
          </div>
        </b-modal>
      </section>

      <!-- SCRY MODAL  -->
      <b-modal :active="isScryModalOpen">
        <div v-if="self" class="modal-card" width="400px">
          <header class="modal-card-head"></header>
          <section v-if="self.Boardstate.Library" class="modal-card-body">
            <!-- TODO: Handle scry X instead of assuming just scry 1 -->
            <Card v-if="isScryModalOpen" v-bind="self.Boardstate.Library[0]" />
          </section>
          <footer class="modal-card-foot">
            <b-button @click="toggleScryModal()">Close</b-button>
            <b-button @click="handleScryBottom()">Bottom</b-button>
          </footer>
        </div>
      </b-modal>
      <!-- END SCRY MODAL  -->
    </section>

    <!-- ### TOOLBAR START  -->
    <template>
      <b-navbar fixed-bottom>
        <template #start>
          <b-navbar-item @click="handleDraw()">
            <button class="button is-primary is-small">
              <span class="icon">
                <i class="fa fa-book"></i>
              </span>
              <span>Draw</span>
            </button>
          </b-navbar-item>
          <b-navbar-item @click="toggleScryModal()" href="#">
            <button class="button is-primary is-small">
              <span class="icon">
                <i class="fa fa-book"></i>
              </span>
              <span>Scry</span>
            </button>
          </b-navbar-item>
          <b-navbar-item>
            <b-button type="is-primary is-small" @click="isInviteModalOpen = !isInviteModalOpen">Invite a
              friend</b-button>
          </b-navbar-item>
        </template>

        <template #end>
          <b-navbar-item tag="div">
            <div class="buttons">
              <a @click="isCommanderSelectionOpen = !isCommanderSelectionOpen"
                class="button is-dark is-small">Commanders</a>
              <!-- <a @click="handleTapAll()" class="button is-dark is-small"><strong>Tap All</strong></a> -->
              <!-- <a @click="handleUntapAll()" class="button is-light is-small">Untap All</a> -->
            </div>
          </b-navbar-item>
        </template>
      </b-navbar>
    </template>
  </div>
</template>
<script>
import _ from 'lodash';
import interact from 'interactjs';
import DraggableCard from '@/components/DraggableCard';
import Card from '@/components/Card';
import Opponent from '@/components/Opponent';
import { mapState } from 'vuex';

export default {
  name: 'board',
  data() {
    return {
      keepFirst: false,
      openOnFocus: false,
      searchQuery: '',
      isStackOpen: true,
      selected: '',
      clearable: true,
      selectedCommander: undefined,
      isCommanderSelectionOpen: true,
      isInviteModalOpen: false,
      isScryModalOpen: false,
      isCreateTokenModalOpen: false,
      handlers: {
        tap: this.handleTap,
        cast: this.handleCast,
      }
    };
  },
  created() {
    // load the initial game state
    this.$store.dispatch('Games/getGame', {
      gameID: this.$route.params.id
    })

    // subscribe to updates
    this.$store.dispatch('Games/subscribeToGame', {
      gameID: this.$route.params.id,
      userID: this.user.ID,
    });

    interact('#Stack').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,
      // listen for drop related events:
      ondropactivate: function (event) {
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });

    interact('#Battlefield').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,
      // listen for drop related events:
      ondropactivate: function (event) {
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });

    interact('#Hand').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,

      // listen for drop related events:
      ondropactivate: function (event) {
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });

    interact('#Graveyard').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,

      // listen for drop related events:
      ondropactivate: function (event) {
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });

    interact('#Exiled').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,

      // listen for drop related events:
      ondropactivate: function (event) {
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });

    interact('#Revealed').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,

      // listen for drop related events:
      ondropactivate: function (event) {
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });
  },
  computed: {
    ...mapState({
      game: (state) => state.Games.game,
      user: (state) => state.Users.User,
    }),
    // self returns the authenticated user's player object from the game's players.
    self() {
      const userID = this.$store.state.Users.User.ID
      const players = this.$store.state.Games.game.Players
      const found = players.find(x => x.ID === userID)
      return found
    },
    // opps returns an array of players that are not the self player
    opps() {
      const players = this.$store.state.Games.game.Players
      const userID = this.$store.state.Users.User.ID
      const opps = players.filter(x => x.ID != userID)
      return opps
    },
    filteredCommanderData() {
      const filteredCommanders = this.self.Boardstate.Library.filter((card) => {
        const name = card.Name.toLowerCase()
        const query = this.searchQuery.toLowerCase()
        return name.includes(query);
      })
      return filteredCommanders
    }
  },
  methods: {
    handleCast(card) {
      let self = this.self;
      let game = this.game;

      // find the card in hand, graveyard, or exile
      for (let zone in self.Boardstate) {
        console.log('searching ', zone)

        // don't loop if it's not an array
        if (!Array.isArray(self.Boardstate[zone])) {
          continue
        }

        // don't loop if it's empty
        if (self.Boardstate[zone].length === 0) {
          continue
        }

        console.log(self.Boardstate[zone])
        
        // not an empty pile, try to find target card
        let idx = self.Boardstate[zone].findIndex(x => x.ID === card.ID)
        if (idx > -1) {
          // move it to stack 
          console.log(self)
          game.Stack.push(self.Boardstate[zone][idx])
          self.Boardstate[zone].splice(idx, 1)
        }
      }
    },
    handleDraw() {
      let self = this.self
      let draw = self.Boardstate.Library.shift()
      self.Boardstate.Hand.push(draw)
      this.handleChange()
    },
    handleChange() {
      this.$store.dispatch('Games/sync', this.game);
    },
    handleTap(card) {
      card.Tapped = !card.Tapped;
      this.handleChange();
    },
    handleTapAll() {
      // this.$store.dispatch('tapAll', this.self);  // TODO
    },
    handleUntapAll() {
      // this.$store.dispatch('untapAll', this.self); // TODO 
    },
    toggleScryModal() {
      this.isScryModalOpen = !this.isScryModalOpen;
    },
    toggleCreateTokenModal() {
      // this.isCreateTokenModalOpen = !this.isCreateTokenModalOpen;
    },
    handleScryBottom() {
      this.isScryModalOpen = false;
      const copy = Object.assign({}, this.self);
      const card = copy.Boardstate.Library.shift();
      if (card) {
        // scrying an empty library causes nothing to happen
        copy.Boardstate.Library.push(card);
        this.handleChange();
      }
    },
    // TODO: pull invite link out into its own component
    inviteLink() {
      const link = `www.edhgo.com/join/${this.$route.params.id}`;
      return link;
    },
    copyToClipboard(text) {
      var dummy = document.createElement('textarea');
      // to avoid breaking orgain page when copying more words
      // cant copy when adding below this code
      // dummy.style.display = 'none'
      document.body.appendChild(dummy);
      //Be careful if you use texarea. setAttribute('value', value),
      // which works with "input" does not work with "textarea"
      dummy.value = text;
      dummy.select();
      document.execCommand('copy');
      document.body.removeChild(dummy);
      this.isInviteModalOpen = false;
    },
    // commander is a card type here
    addCommander(commander) {
      this.self.Boardstate.Library = this.self.Boardstate.Library.filter((option) => (option ? commander.Name != option.Name : false));
      this.self.Boardstate.Commander.push(commander);
      this.handleChange();
    },
    removeCommander(commander) {
      this.self.Boardstate.Library.push(commander);
      this.self.Boardstate.Commander = this.self.Boardstate.Commander.filter((option) => commander.Name != option.Name);
      this.handleChange();
    },
    triggerShuffle() {
      // todo 
    }
  },
  components: {
    Opponent,
    DraggableCard,
  },
};
</script>
<style media="screen" scoped>
#Stack {
  display: flex;
  flex-wrap: row;
  height: 500px;
}

#Battlefield {
  display: flex;
  width: auto;
  height: 500px;
}

#Graveyard {
  display: flex;
  height: auto;
  width: auto;
  margin: 10px;
}

#Exiled {
  display: flex;
  height: auto;
  width: auto;
  margin: 10px;
}

#Revealed {
  display: flex;
  height: auto;
  width: auto;
  margin: 10px;
}

.shell {
  padding: 0.5rem;
  border: 1px solid #efefef;
  margin: 0.25rem 0rem;
}

.hand {
  height: 170px;
  display: flex;
  flex-wrap: row;
}

.bordered {
  border: 1px #000;
}

/* 
 * InteractJS styles 
*/
.outer-dropzone {
  background-color: #e3c5ff;
  min-height: 250px;
  height: auto;
}

.inner-dropzone {
  background-color: #e3c5ff;
  height: 80px;
}

.dropzone {
  background-color: #e3c5ff;
  border: dashed 4px transparent;
  border-radius: 4px;
  margin: 10px auto 30px;
  padding: 10px;
  width: 97vw;
  transition: background-color 0.3s;
}

.drop-active {
  border-color: #aaa;
}

.drop-target {
  background-color: #29e;
  border-color: #fff;
  border-style: solid;
}

.drag-drop {
  display: inline-block;
  padding: 2em 0.5em;
  margin: 1rem 0 0 1rem;
  border: solid 2px #fff;
  touch-action: none;
  transform: translate(0px, 0px);
  transition: background-color 0.3s;
}

.drag-drop.can-drop {
  transition: background-color 0.3s;
  margin: 1rem 0 0 1rem;
  border: solid 2px #fff;
}

.container {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-around;
  align-items: center;
}

.item {
  position: relative;
  margin: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
}
</style>
