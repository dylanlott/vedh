<template>
  <div class="container">
    <div class="columns is-mobile is-centered">
      <div class="column is-half-desktop is-full-mobile" v-if="user">
          <div class="box decklist">
        <h1 class="title is-1">Hi, {{ user.Username }}.</h1>
        <h1 class="title is-3">Start a game</h1>
            <p>1. Copy a decklist from <a href="https://www.archidekt.com">Archidekt</a>. <i>Make sure to select CSV format when you export it.</i></p>
            <br>
            <p>2. Paste your decklist here.</p>
            <br>
            <b-field label="Decklist" :label-position="labelPosition">
              <b-input v-model="decklist" maxlength="20000" type="textarea"></b-input>
            </b-field>
            <b-button type="is-success" expanded @click="handleCreateGame">Go</b-button>
          </div>
      </div>
    </div>
    <!-- MTG JSON Credit -->
    <div class="columns is-mobile is-centered">
      <a href="https://mtgjson.com" style="display: inline-flex; align-items: center;">
        <!-- <img 
          TECHDEBT fix this image
          src="http://mtgjson.com/images/assets/logo-mtgjson-light-blue.svg" 
          width="40px" 
          title="MTGJSON logo"> -->
        <p class="is-size-6" style="margin-left: 10px">Powered by MTGJSON</p>
      </a>
    </div>
  </div>
</template>
<script>
import { mapState } from 'vuex';

export default {
  name: 'game',
  data() {
    return {
      isFullPage: false,
      labelPosition: 'on-border',
      decklist: '',
    };
  },
  computed: {
    ...mapState({
      user: (state) => state.Users.User,
      loading: state => state.Games.loading,
    }),
  },
  methods: {
    handleCreateGame() {
      let g = { 
        ID: '',
        Turn: {
          Player: this.user.Username,
          Phase: 'setup',
          Number: 0,
        },
        Players: [{
          GameID: '',
          User: this.user.Username,
          UserID: this.user.ID,
          Life: 40,
          Commander: [],
          Library: [],
          Decklist: this.decklist,
          Graveyard: [],
          Exiled: [],
          Battlefield: [],
          Hand: [],
          Revealed: [],
          Controlled: [],
        }],
      }
      this.$store.dispatch('Games/createGame', g)
    },
  },
};
</script>
<style>
.shell {
  margin: 0.5rem;
}
.decklist {
  margin: 0.5rem;
}
</style>