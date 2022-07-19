<template>
  <div class="container is-fluid" v-if="user && game">
    <b-button type="is-primary" @click="isInviteModalOpen = !isInviteModalOpen">Invite a friend</b-button> 
    <b-modal :active="isInviteModalOpen">
      <div v-if="self" class="modal-card" width="400px">
        <header class="modal-card-head">Get an Invite Link</header>
        <section class="modal-card-body">
          <code id="invite-link">{{ inviteLink() }}</code> 
          <b-button type="is-primary" @click="copyToClipboard(inviteLink())">Copy</b-button> 
        </section>
      </div>
    </b-modal>
    <!-- TURN TRACKER -->
    <!-- <div class="box"> -->
      <!-- <TurnTracker :game="game"/> -->
    <!-- </div> -->
    <!-- END TURN TRACKER -->

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
    
    <!-- CREATE TOKEN MODAL  -->
    <!-- TODO: Implement the create token modal.
    <b-modal :active="isCreateTokenModalOpen">
      <div v-if="self" class="modal-card" width="400px">
        <header class="modal-card-head"></header>
        <section class="modal-card-body">
         Create Token
         Name:  
         Type:   
         Power:   <b-field>
      <b-numberinput v-model="power"></b-numberinput>
      </b-field>
         Toughness: 
          <b-field>
      <b-numberinput v-model="toughness"></b-numberinput>
    </b-field>
        </section>
        <footer class="modal-card-foot">
          <b-button @click="toggleCreateTokenModal()">Close</b-button>
        </footer>
      </div>
    </b-modal> -->
    <!-- END CREATE TOKEN MODAL  -->

    <!-- OPPONENTS BOARDSTATES -->
    <div v-if="bs.length > 0">
      <div class="box" :key="player.ID" v-for="player in bs">
        <pre>{{player}}</pre>
      <div class="columns" v-if="user.ID !== player.User.ID">
        <div class="title">{{ player.User.Username }}</div>
        <div class="battlefield">
          <div class="columns" v-if="player">
            <div class="column">
              <draggable
                class="columns is-mobile"
                v-model="player.Field"
                group="people"
                @start="drag = true"
                @end="drag = false"
              >
                <div class="column" v-for="card in player.Field" :key="card.id">
                  <Card v-bind="card"></Card>
                </div>
              </draggable>
            </div>
          </div>
        </div>
      </div>
    </div>
    </div>
    <!-- END OF OPPONENTS BOARDSTATES -->

    <!-- SELF BOARDSTATE - PUBLIC SECTION -->
    <p class="title is-6">Battlefield</p>
    <div class="columns is-mobile is-desktop is-flex" id="selfBattlefield" v-if="self">
      <!-- SELF - BATTLEFIELD -->
      <div class="column is-desktop is-mobile box is-flex">
        <draggable
          class="columns is-flex is-multiline is-mobile is-align-items-flex-start"
          @change="handleChange()"
          v-model="self.Field"
          group="people"
          @start="drag = true"
          @end="drag = false"
        >
          <!-- Note: Cards can only be tapped on the battlefield -->
          <div v-on:dblclick="handleTap(card)" 
            class="column" 
            v-for="card in self.Field" 
            :key="card.id">
              <Card v-bind="card"></Card>
          </div>
        </draggable>
      </div>
    </div>
    <!-- END SELF BATTLEFIELD -->

    <!-- ### TOOLBAR START  -->
    <template>
      <b-navbar>
        <template #start>
          <b-navbar-item @click="handleDraw()" href="#">
            <button class="button is-primary">
              <span class="icon">
                <i class="fa fa-book"></i>
              </span>
              <span>Draw</span>
            </button>
          </b-navbar-item>
          <b-navbar-item @click="toggleScryModal()" href="#">
            <button class="button is-primary">
              <span class="icon">
                <i class="fa fa-book"></i>
              </span>
              <span>Scry</span>
            </button>
          </b-navbar-item>
        </template>

        <template #end>
          <b-navbar-item tag="div">
            <div class="buttons">
              <a @click="handleTapAll()" class="button is-dark "><strong>Tap All</strong></a>
              <a @click="handleUntapAll()" class="button is-light">Untap All</a>
              <!-- <a @click="toggleCreateTokenModal" class="button is-primary">Create Token</a> -->
            </div>
          </b-navbar-item>
        </template>
      </b-navbar>
    </template>

    <!-- FOOTER  -->
    <footer class="footer">
      <div class="content">
        <div class="column is-full">
          <!-- SELF - HAND-->
          <p class="title is-5">Hand</p>
          <draggable
            class="columns is-flex is-multiline is-mobile is-align-items-flex-start"
            v-model="self.Hand"
            group="people"
            @change="handleChange()"
            @start="drag = true"
            @end="drag = false"
          >
            <div class="column" v-for="card in self.Hand" :key="card.id">
              <Card v-bind="card"></Card>
            </div>
          </draggable>
        </div>
      </div>
    </footer>
  </div>
</template>
<script>
import _ from 'lodash';
import draggable from 'vuedraggable';
import Card from '@/components/Card';
import { mapState } from 'vuex';

export default {
  name: 'board',
  data() {
    return {
      isInviteModalOpen: false,
      isScryModalOpen: false,
      isCreateTokenModalOpen: false,
    };
  },
  created() {
    // load and sub to game
    // this.$store.dispatch('getGame', this.$route.params.id)
    this.$store.dispatch('subscribeToGame', {
      gameID: this.$route.params.id,
      userID: this.user.ID,
    });
    // load and sub to all boardstates
    this.$store.dispatch('subAllBoardstates', {
      gameID: this.$route.params.id,
      obsID: this.user.ID,
    });
  },
  computed: mapState({
    game: (state) => state.Games.game,
    bs: (state) => state.Boardstates.boardstates,
    self: (state) => state.Boardstates.self,
    user: (state) => state.Users.User,
  }),
  methods: {
    handleDraw() {
      this.$store.dispatch('draw', this.self);
    },
    handleChange() {
      this.$store.dispatch('mutateBoardState', this.self);
    },
    handleTap(card) {
      // TODO: Make this a vuex boardstate action
      card.Tapped = !card.Tapped;
      this.handleChange();
    },
    handleTapAll() {
      this.$store.dispatch('tapAll', this.self);
    },
    handleUntapAll() {
      this.$store.dispatch('untapAll', this.self);
    },
    toggleScryModal() {
      this.isScryModalOpen = !this.isScryModalOpen;
    },
    toggleCreateTokenModal() {
      this.isCreateTokenModalOpen = !this.isCreateTokenModalOpen
    },
    handleScryBottom() {
      this.isScryModalOpen = false
      const copy = Object.assign({}, this.self)
      const card = copy.Library.shift()
      if (card) {
        // scrying an empty library causes nothing to happen
        copy.Library.push(card)
        return this.$store.dispatch('mutateBoardState', copy)
      } 
    },
    inviteLink() {
      const link = `www.edhgo.com/join/${this.$route.params.id}`
      return link
    },
    copyToClipboard(text) {
      var dummy = document.createElement("textarea");
      // to avoid breaking orgain page when copying more words
      // cant copy when adding below this code
      // dummy.style.display = 'none'
      document.body.appendChild(dummy);
      //Be careful if you use texarea. setAttribute('value', value), 
      // which works with "input" does not work with "textarea"
      dummy.value = text;
      dummy.select();
      document.execCommand("copy");
      document.body.removeChild(dummy);
      alert("copied text ", text)
    }
  },
  components: {
    draggable,
    Card,
  },
};
</script>
<style media="screen" scoped>
.shell {
  padding: 0.5rem;
  border: 1px solid #efefef;
  margin: 0.25rem 0rem;
}

.bordered {
  border: 1px #000;
}
</style>
