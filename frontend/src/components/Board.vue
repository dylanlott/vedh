<template>
  <div class="container is-fluid" v-if="user && game">
    <section>
      <b-button type="is-primary is-small" @click="isInviteModalOpen = !isInviteModalOpen">Invite a friend</b-button>
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

    <!-- TURN TRACKER -->
    <!-- <div class="box"> -->
    <!-- <TurnTracker :game="game"/> -->
    <!-- </div> -->
    <!-- END TURN TRACKER -->

    <!-- COMMANDER SELECTION  -->
    <section>
      <b-modal has-modal-card :active="isCommanderSelectionOpen" trap-focus :destroy-on-hide="false" aria-role="dialog"
        aria-label="Example Modal" close-button-aria-label="Close" aria-modal>
        <div class="card">
          <div class="card-content">
            <div class="content">
              <b-field label="Select your Commander(s) from your library:">
                <b-autocomplete v-model="name" placeholder="e.g. Kykar, Wind's Fury" :open-on-focus="openOnFocus"
                  :data="filteredCommanderData" field="Name"
                  @select="option => { option ? addCommander(option) : null }" :clearable="clearable">
                </b-autocomplete>
              </b-field>
              <b-field>
                <b-tag :key="key" v-if="self.Commander.length > 0" v-for="(obj, key) in self.Commander"
                  type="is-primary" closable aria-close-label="Close tag" @close="removeCommander(obj)">
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
    <div class="box">
      <div v-for="player in bs" v-if="player.User.ID !== user.ID">
        <p class="title is-6">{{ player.User.Username }}</p>
        <div class="tile" :key="player.ID" v-if="player.User.ID !== user.ID" v-for="player in bs">
          <div class="columns">
            <div class="" v-for="card in player.Field">
              <!-- TODO: make this prettier for opponent views -->
              <Card v-bind="card"></Card>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- SELF BOARDSTATE - PUBLIC SECTION -->
    <p class="title is-6">Your Battlefield</p>
    <div class="columns is-mobile is-desktop is-flex" id="selfBattlefield" v-if="self">
      <!-- SELF - BATTLEFIELD -->
      <div class="column is-desktop is-mobile box is-flex">
        <draggable class="columns is-flex is-multiline is-mobile is-align-items-flex-start" @change="handleChange()"
          v-model="self.Field" group="people" @start="drag = true" @end="drag = false">
          <!-- Note: Cards can only be tapped on the battlefield -->
          <div v-on:dblclick="handleTap(card)" class="column" v-for="card in self.Field" :key="card.id">
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
        </template>

        <template #end>
          <b-navbar-item tag="div">
            <div class="buttons">
              <a @click="isCommanderSelectionOpen = !isCommanderSelectionOpen"
                class="button is-dark is-small">Commanders</a>
              <a @click="handleTapAll()" class="button is-dark is-small"><strong>Tap All</strong></a>
              <a @click="handleUntapAll()" class="button is-light is-small">Untap All</a>
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
          <draggable class="columns is-flex is-multiline is-mobile is-align-items-flex-start" v-model="self.Hand"
            group="people" @change="handleChange()" @start="drag = true" @end="drag = false">
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
      // TODO get rid of soon
      keepFirst: false,
      openOnFocus: false,
      name: '',
      selected: '',
      clearable: true,
      isCommanderSelectionOpen: true, // NB: default open at start

      isInviteModalOpen: false,
      isScryModalOpen: false,
      isCreateTokenModalOpen: false,
    };
  },
  created() {
    this.$store.dispatch('subscribeToGame', {
      gameID: this.$route.params.id,
      userID: this.user.ID,
    });
    this.$store.dispatch('subAllBoardstates', {
      gameID: this.$route.params.id,
      obsID: this.user.ID,
    });
  },
  computed: {
    // TODO this belongs with Commander Selection modal in its own component
    filteredCommanderData() {
      if (this.self.Library && this.self.Library.length > 0) {
        return this.self.Library.filter(option => {
          return (
            option.Name.toString().toLowerCase().indexOf(this.name.toLowerCase()) >= 0
          )
        })
      }
    },
    ...mapState({
      game: (state) => state.Games.game,
      bs: (state) => state.Boardstates.boardstates,
      self: (state) => state.Boardstates.self,
      user: (state) => state.Users.User,
    })
  },
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
      this.isCreateTokenModalOpen = !this.isCreateTokenModalOpen;
    },
    handleScryBottom() {
      this.isScryModalOpen = false;
      const copy = Object.assign({}, this.self);
      const card = copy.Library.shift();
      if (card) {
        // scrying an empty library causes nothing to happen
        copy.Library.push(card);
        return this.$store.dispatch('mutateBoardState', copy);
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
      this.isInviteModalOpen = false
    },
    // commander is a card type here 
    addCommander(commander) {
      this.self.Library = this.self.Library.filter((option) => option ? commander.Name != option.Name : false)
      this.self.Commander.push(commander)
      this.handleChange()
    },
    removeCommander(commander) {
      this.self.Library.push(commander)
      this.self.Commander = this.self.Commander.filter((option) => commander.Name != option.Name)
      this.handleChange()
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
