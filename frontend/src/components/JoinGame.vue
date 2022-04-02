<template>
  <div class="join shell">
    <div class="container">
      <div class="columns is-mobile is-centered">
        <div class="column is-half">
          <h1 class="title is-1">You've been invited to a Commander game.</h1>
          <h1 class="title is-4">Paste a decklist and join.</h1>

          <p v-if="game.PlayerIDs.length > 1">There are {{ game.PlayerIDs.length }} other players in this game.</p>
          <p v-if="game.PlayerIDs.length === 1">There is {{ game.PlayerIDs.length }} other player in this game.</p>
          <p v-if="game.PlayerIDs.length === 0">There is no other player in this game. Are you sure you got the code right?</p><br>

          <b-field label="Decklist" label-position="on-border">
            <b-input maxlength="200000" v-model="decklist" type="textarea"></b-input>
          </b-field>

          <div v-if="!user.Username">
            <b-field label="Add a username?">
              <b-input v-model="username"></b-input>
            </b-field>
          </div>

          <b-button @click="handleJoinGame()" type="button" class="is-success">Join Game</b-button>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import { mapState } from 'vuex';

export default {
  name: 'join',
  data() {
    return {
      decklist: '',
      username: '',
    };
  },
  computed: {
    ...mapState({
      game: (state) => state.Games.game,
      user: (state) => state.Users.User,
    }),
  },
  created() {
    this.$store.dispatch('getGame', this.$route.params.id);
  },
  methods: {
    handleJoinGame() {
      var rid = this.uuid()
      this.$store.dispatch('joinGame', {
        inputGame: {
          ID: this.$route.params.id,
          Decklist: this.decklist,
          User: {
            ID: this.user.ID || rid,
            Username: this.user.Username || this.username,
          },
          BoardState: {
            GameID: this.$route.params.id,
            User: {
              Username: this.user.Username || this.username,
              ID: this.user.ID || rid,
            },
            Life: 40,
            Commander: [],
          },
        },
      });
    },
    uuid() {
      return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
        var r = (Math.random() * 16) | 0,
          v = c == 'x' ? r : (r & 0x3) | 0x8;
        return v.toString(16);
      });
    },
  },
};
</script>
<style>
</style>