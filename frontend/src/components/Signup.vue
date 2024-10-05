<template>
  <section class="shell">
    <div class="columns is-centered">
      <div class="column is-4 is-mobile">
        <div class="box">
          <h1 class="title is-1">Signup for EDH-Go</h1>
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
          <b-button @submit="handleSignup()" @click="handleSignup()" 
          v-on:keyup.enter="handleSignup()" type="is-primary">
            Sign Up
          </b-button>
          <div class="already-a-member">
            Already a member? <a href="/login">Login.</a>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script>
export default {
  name: 'signup',
  data: () => {
    return {
      labelPosition: 'on-border',
      username: '',
      password: '',
    };
  },
  methods: {
    handleSignup() {
      this.$store.dispatch('Users/signup', {
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
    },
  },
};
</script>

<style>
.already-a-member {
  margin-top: 15px
}
</style>