<template>
  <section class="shell">
    <div class="columns is-centered">
      <div class="column is-4 is-mobile">
        <form @submit.prevent="handleLogin">
          <div class="box">
            <h1 class="title is-1">Login to vEDH</h1>
            <b-field label="Username" :label-position="labelPosition">
              <b-input v-model="username"></b-input>
            </b-field>
            <b-field
              v-on:keyup.enter="handleLogin()"
              @submit="handleLogin()"
              label="Password"
              :label-position="labelPosition"
            >
              <b-input type="password" v-model="password"></b-input>
            </b-field> 
    
            <!--  LOADING BAR -->
            <b-progress v-if="isLoading"></b-progress>

            <b-button native-type="submit" @click="handleLogin()" v-on:keyup.enter="handleLogin()" type="is-primary">
              Login
            </b-button>
            <div class="not-a-member">Not a member? <a href="/signup">Sign up.</a></div>
          </div>
        </form>
      </div>
    </div>
  </section>
</template>
<script>
export default {
  name: 'login',
  data() {
    return {
      username: '',
      password: '',
      labelPosition: 'on-border',
      isLoading: false,
      isFullPage: true,
    };
  },
  computed: {
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
    handleLogin() {
      if (this.isInputValid) {
        this.isLoading = true;
        this.$store
          .dispatch('Users/login', {
            username: this.username,
            password: this.password,
          })
          .then(() => {
            this.username = '';
            this.password = '';
          })
          .catch((err) => {
            this.username = '';
            this.password = '';
          });
      }
    },
  },
};
</script>

<style>
.form {
  max-width: 480px;
  margin: 0 auto;
}

.not-a-member {
  margin-top: 15px;
}
</style>
