<template>
  <div
    v-bind:class="{
      'd-none': hidden,
      tapped: Tapped,
      flipped: flipped,
    }"
  >
    <div class="card fixed-size">
      <img :src="`https://api.scryfall.com/cards/${ScryfallID}?format=image`">
    </div>
  </div>
</template>
<script>
export default {
  name: 'Card',
  data() {
    return {
      hidden: false, // if a card can be seen at all - visibility off
      flipped: false, // if a card is upside down or not
      trackers: {}, // player-assigned trackers
      labels: {}, // player assigned labels
      counters: {}, // game-assigned counters such as poison or infect
      reminders: {}, // untap effects, etb effect reminders, etc...
    };
  },
  props: [
    'ID',
    'Name',
    'CMC',
    'ManaCost',
    'colorIdentity',
    'Power',
    'Toughness',
    'Text',
    'Types',
    'Supertypes',
    'Subtypes',
    'ScryfallID',
    'Tapped',
    'Flipped',
  ],
};
</script>
<style scoped>
.fixed-size {
  width: 275px;
  max-width: 275px;
}
.flipped {
  color: #000;
}
.column {
  padding: 0rem;
}
.card-content {
  padding: 0.7rem;
}
</style>
