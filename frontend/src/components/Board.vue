<template>
  <section>

    <!-- DEBUG UTILITIES - UNCOMMENT TO SEE THIS DATA -->
    <!-- <code>  
      {{ self }}
    </code>
    <br>
    <code>  
      {{  game }}
    </code> -->

    <section>
      <!-- ### BATTLEFIELD -->
      <section v-if="self" id="Battlefield" class="battlefield dropzone outer-dropzone">
        BATTLEFIELD
        <DraggableCard v-for="card in self.Boardstate.Field" :card="card" :user="user" :key="card.ID" />
      </section>

      <section id="Hand" v-if="self" class="hand dropzone outer-dropzone container">
        HAND
        <DraggableCard class="item" v-for="card in self.Boardstate.Hand" :card="card" :user="user" :key="card.ID" />
      </section>

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
      <!-- <section>
        <b-modal
          has-modal-card
          :active="isCommanderSelectionOpen"
          trap-focus
          :destroy-on-hide="false"
          aria-role="dialog"
          aria-label="Example Modal"
          close-button-aria-label="Close"
          aria-modal
        >
          <div class="card">
            <div class="card-content">
              <div class="content">
                <b-field label="Select your Commander(s) from your library:">
                  <b-autocomplete
                    v-model="name"
                    placeholder="e.g. Kykar, Wind's Fury"
                    :open-on-focus="openOnFocus"
                    :data="filteredCommanderData"
                    field="Name"
                    @select="
                      (option) => {
                        option ? addCommander(option) : null;
                      }
                    "
                    :clearable="clearable"
                  >
                  </b-autocomplete>
                </b-field>
                <b-field>
                  <b-tag
                    :key="key"
                    v-if="self.Commander.length > 0"
                    v-for="(obj, key) in self.Commander"
                    type="is-primary"
                    closable
                    aria-close-label="Close tag"
                    @close="removeCommander(obj)"
                  >
                    <b>{{ obj.Name }}</b>
                  </b-tag>
                </b-field>
              </div>
            </div>
            <footer class="modal-card-foot">
              <b-button
                type="is-primary is-light"
                size="is-small"
                @click="isCommanderSelectionOpen = !isCommanderSelectionOpen"
                >Done</b-button
              >
            </footer>
          </div>
        </b-modal>
      </section> -->

      <!-- SCRY MODAL  -->
      <b-modal :active="isScryModalOpen">
        <div v-if="self" class="modal-card" width="400px">
          <header class="modal-card-head"></header>
          <section v-if="self.Library" class="modal-card-body">
            <!-- TODO: Handle scry X instead of assuming just scry 1 -->
            <Card v-if="isScryModalOpen" v-bind="self.Library[0]" />
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
            <b-button type="is-primary is-small" @click="isInviteModalOpen = !isInviteModalOpen"
              >Invite a friend</b-button
            >
          </b-navbar-item>
        </template>

        <template #end>
          <b-navbar-item tag="div">
            <div class="buttons">
              <a @click="isCommanderSelectionOpen = !isCommanderSelectionOpen" class="button is-dark is-small"
                >Commanders</a
              >
              <a @click="handleTapAll()" class="button is-dark is-small"><strong>Tap All</strong></a>
              <a @click="handleUntapAll()" class="button is-light is-small">Untap All</a>
            </div>
          </b-navbar-item>
        </template>
      </b-navbar>
    </template>
  </section>
</template>
<script>
import _ from 'lodash';
import interact from 'interactjs';
import DraggableCard from '@/components/DraggableCard';
import Card from '@/components/Card';
import { mapState } from 'vuex';

export default {
  name: 'board',
  data() {
    return {
      // TODO get rid of soon
      keepFirst: false,
      openOnFocus: false,
      name: '',
      selected: '',
      clearable: true,
      isCommanderSelectionOpen: true, // NB: default open at start
      // modal bools
      isInviteModalOpen: false,
      isScryModalOpen: false,
      isCreateTokenModalOpen: false,
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

    // enable draggables to be dropped into this
    interact('#Battlefield').dropzone({
      // Require a 50% element overlap for a drop to be possible
      overlap: 0.50,
      // listen for drop related events:
      ondropactivate: function (event) {
        // console.log('ON DRAG ACTIVATE BATTLEFIELD', event);
        // add active dropzone feedback
        event.target.classList.add('drop-active');
      },
      ondragenter: function (event) {
        // console.log('ON DRAG ENTER BATTLEFIELD', event);
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
        // draggableElement.textContent = 'Dragged in';
      },
      ondragleave: function (event) {
        // console.log('ON DRAG LEAVE BATTLEFIELD', event);
        // remove the drop feedback style
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
        // event.relatedTarget.textContent = 'Dragged out';
      },
      ondrop: function (event) {
        console.log("dropped into battlefield:", event)
      },
      ondropdeactivate: function (event) {
        // console.log('ON DROP DEACTIVATE', event);
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
        // console.log('ON DRAG ENTER HAND', event);
        var draggableElement = event.relatedTarget;
        var dropzoneElement = event.target;

        // feedback the possibility of a drop
        dropzoneElement.classList.add('drop-target');
        draggableElement.classList.add('can-drop');
        // draggableElement.textContent = 'Dragged in';
      },
      ondragleave: function (event) {
        // remove the drop feedback style
        // console.log('ON DRAG LEAVE HAND', event);
        event.target.classList.remove('drop-target');
        event.relatedTarget.classList.remove('can-drop');
      },
      ondrop: function (event) {
        console.log("dropped into hand: ", event)
      },
      ondropdeactivate: function (event) {
        // remove active dropzone feedback
        event.target.classList.remove('drop-active');
        event.target.classList.remove('drop-target');
      },
    });  
  },
  computed: {
    // TECHDEBT this belongs with Commander Selection modal in its own component
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
    }
  },
  methods: {
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
      // TODO: Make this a vuex boardstate action
      // card.Tapped = !card.Tapped;
      // this.handleChange();
    },
    handleTapAll() {
      // this.$store.dispatch('tapAll', this.self);
    },
    handleUntapAll() {
      // this.$store.dispatch('untapAll', this.self);
    },
    toggleScryModal() {
      // this.isScryModalOpen = !this.isScryModalOpen;
    },
    toggleCreateTokenModal() {
      // this.isCreateTokenModalOpen = !this.isCreateTokenModalOpen;
    },
    handleScryBottom() {
      // this.isScryModalOpen = false;
      // const copy = Object.assign({}, this.self);
      // const card = copy.Library.shift();
      // if (card) {
      //   // scrying an empty library causes nothing to happen
      //   copy.Library.push(card);
      //   return this.$store.dispatch('mutateBoardState', copy);
      // }
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
      // this.self.Library = this.self.Library.filter((option) => (option ? commander.Name != option.Name : false));
      // this.self.Commander.push(commander);
      // this.handleChange();
    },
    removeCommander(commander) {
      // this.self.Library.push(commander);
      // this.self.Commander = this.self.Commander.filter((option) => commander.Name != option.Name);
      // this.handleChange();
    },
  },
  components: {
    Card,
    DraggableCard,
  },
};
</script>
<style media="screen" scoped>
#Battlefield {
  display: flex;
  height: 500px;
}
.shell {
  padding: 0.5rem;
  border: 1px solid #efefef;
  margin: 0.25rem 0rem;
}
.battlefield {
  border: 1px black;
}
.hand {
  height: 170px;
  display: flex;
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
  border: 1px solid black;
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
