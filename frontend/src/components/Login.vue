<template>
  <section class="shell">
    <div class="columns is-centered">
      <div class="column is-4 is-mobile">
        <div class="box">
          <h1 class="title is-1">Login to EDH-Go</h1>
          <b-field label="Username" :label-position="labelPosition">
            <b-input v-model="username"></b-input>
          </b-field>
          <b-field
            v-on:keyup.enter="handleSignup()"
            @submit="handleSignup()"
            label="Password"
            :label-position="labelPosition">
            <b-input type="password" v-model="password"></b-input>
          </b-field>
          <b-button @submit="onLoginClick()" @click="onLoginClick()" 
          v-on:keyup.enter="onLoginClick()" type="is-primary">
            Login
          </b-button>
          <div class="not-a-member">Not a member? <a href="/signup">Sign up.</a></div>
        </div>
      </div>
    </div>
  </section>
</template>
<script>
import { mapState } from 'vuex';
export default {
  data() {
    return {
      username: '',
      password: '',
      labelPosition: "on-border",
    };
  },
  computed: {
    ...mapState({
      loading: (state) => state.User.loading,
    }),
    isInputValid() {
      if (this.username.length === 0) {
        return false;
      }
      if (this.password.length === 0) {
        return false;
      }
      return true;
    },
  },
  methods: {
    onLoginClick() {
      if (this.isInputValid) {
        this.$store.dispatch('login', {
          username: this.username,
          password: this.password,
        })
        .then(() => {
          this.username = ""
          this.password = ""
        })
        .catch((err) => {
          this.username = ""
          this.password = ""
        })
      }
    },
  },
};
</script>

<style lang="scss" scoped>
.form {
  max-width: 480px;
  margin: 0 auto;
}

.not-a-member {
  margin-top: 15px;
}
</style>
