<template>
  <section>
    <Grid />
    <DraggableCard />
  </section>
</template>
<script>
import _ from 'lodash';
import Grid from '@/components/Grid';
import DraggableCard from '@/components/DraggableCard';
import Card from '@/components/Card';
import { mapState } from 'vuex';

export default {
  name: 'board',
  created() {
    this.$store.dispatch('getGame', { 
      gameID: this.$route.params.id
    })
    this.$store.dispatch('subscribeToGame', {
      gameID: this.$route.params.id,
      userID: this.user.ID,
    });
  },
  computed: {
    // TECHDEBT this belongs with Commander Selection modal in its own component
    filteredCommanderData() {
      if (this.self.Library && this.self.Library.length > 0) {
        return this.self.Library.filter((option) => {
          return option.Name.toString().toLowerCase().indexOf(this.name.toLowerCase()) >= 0;
        });
      }
    },
    ...mapState({
      game: (state) => state.Games.game,
      players: (state) => state.Games.game.Players,
      user: (state) => state.Users.User,
    }),
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
      this.isInviteModalOpen = false;
    },
    // commander is a card type here
    addCommander(commander) {
      this.self.Library = this.self.Library.filter((option) => (option ? commander.Name != option.Name : false));
      this.self.Commander.push(commander);
      this.handleChange();
    },
    removeCommander(commander) {
      this.self.Library.push(commander);
      this.self.Commander = this.self.Commander.filter((option) => commander.Name != option.Name);
      this.handleChange();
    },
  },
  components: {
    Grid,
    DraggableCard,
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
