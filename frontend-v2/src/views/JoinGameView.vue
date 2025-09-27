<template>
  <section class="join-game">
    <h1>Join game</h1>
    <p v-if="!gameID">Provide a valid invite link to join a game.</p>
    <form v-else @submit.prevent="handleJoin">
      <p>You are about to join game <strong>{{ gameID }}</strong>.</p>
      <button class="primary" :disabled="games.loading">{{ games.loading ? 'Joining…' : 'Join game' }}</button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useGamesStore } from '../stores/games';
import { useAuthStore } from '../stores/auth';

const route = useRoute();
const router = useRouter();
const games = useGamesStore();
const auth = useAuthStore();

const gameID = computed(() => route.params.id as string | undefined);

async function handleJoin() {
  if (!gameID.value || !auth.profile) return;
  const payload = {
    ID: gameID.value,
    Decklist: '',
    BoardState: {
      UserID: auth.profile.ID,
      User: auth.profile.Username,
      GameID: gameID.value,
      Life: 40,
      Commander: [],
      Library: [],
      Graveyard: [],
      Exiled: [],
      Battlefield: [],
      Hand: [],
      Revealed: [],
      Controlled: [],
      Counters: [],
    },
  } as const;
  const joinedID = await games.joinGame(payload);
  if (joinedID) {
    router.push({ name: 'board', params: { id: joinedID } });
  }
}
</script>

<style scoped lang="scss">
.join-game {
  max-width: 480px;
  margin: 3rem auto;
  padding: 2.5rem;
  border-radius: 20px;
  background: rgba(15, 19, 26, 0.85);
  border: 1px solid rgba(255, 255, 255, 0.08);
}

button.primary {
  border: none;
  border-radius: 10px;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  font-weight: 600;
  background: linear-gradient(120deg, #85d7ff, #3f8cff);
  color: #0b1016;
  cursor: pointer;
}
</style>
