<template>
  <div>
    <!-- Page Content -->
    <div class="shell landing">
      <!-- Jumbotron Header -->
      <section class="hero is-medium is-primary is-bold welcome">
        <div class="hero-body">
          <div class="container">
            <h1 class="title is-1 has-text-centered">
              Sign up below to play Commander with vEDH.
            </h1>
            <div class="columns is-centered">
              <div class="column is-4 is-mobile">
                <div class="box">
                  <b-field label="Username" :label-position="labelPosition">
                    <b-input v-model="username"></b-input>
                  </b-field>
                  <b-field v-on:keyup.enter="handleSignup()" @submit="handleSignup()" label="Password"
                    :label-position="labelPosition">
                    <b-input type="password" v-model="password"></b-input>
                  </b-field>
                  <b-button @submit="handleSignup()" @click="handleSignup()" v-on:keyup.enter="handleSignup()"
                    type="is-primary">
                    Sign Up
                  </b-button>
                  <div class="already-a-member">
                    Already a member? <a href="/login">Login.</a>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- Benefits Tiles -->
      <section class="columns benefits">
        <div class="column">
          <article class="tile is-child notification is-white has-text-centered">
            <h3><i class="fa fa-mobile-alt fa-3x"></i></h3>
            <p class="title">Mobile First</p>
            <p class="subtitle">Play anytime, anywhere, on any device with an internet connection and a browser.</p>
          </article>
        </div>
        <div class="column">
          <article class="tile is-child notification is-white has-text-centered">
            <h3><i class="fa fa-arrows-alt fa-3x"></i></h3>
            <p class="title">Flexible</p>
            <p class="subtitle">Our boardstate engine allows for any format to be played, with a focus on EDH games.</p>
          </article>
        </div>
        <div class="column">
          <article class="tile is-child notification is-white has-text-centered">
            <h3><i class="fa fa-vial fa-3x"></i></h3>
            <p class="title">Simple</p>
            <p class="subtitle">We worked hard to keep the interface intuitive and easy to use on mobile <i>and</i>
              desktop.</p>
          </article>
        </div>
        <div class="column">
          <article class="tile is-child notification is-white has-text-centered">
            <h3><i class="fa fa-tachometer-alt fa-3x"></i></h3>
            <p class="title">Fast</p>
            <p class="subtitle">The app is built from the ground up from scratch to be as small, fast, and reliable as
              possible.</p>
          </article>
        </div>
      </section>

      <!-- Mailchimp signup -->
      <div class="columns is-centered">
        <div class="box">
          <div class="column is-4 is-mobile">
            <Mailchimp />
          </div>
        </div>
      </div>

      <!-- Footer -->
      <footer class="footer">
        <div class="content has-text-centered">
          <p class="text is-size-7"><strong>vEDH.xyz</strong> - powered by the Hivemind.</p>
          <p class="text is-size-7">Build v{{ version }} </p>
        </div>
      </footer>
    </div>
  </div>
</template>
<script type="text/javascript">
import pkg from '../../package.json'
import Mailchimp from './Mailchimp.vue';
import Signup from './Signup.vue';
export default {
  name: 'landing',
  data() {
    return {
      username: "",
      password: "",
      year: new Date().getFullYear().toString(),
      version: pkg.version,
      labelPosition: "centered"
    };
  },
  components: { Signup, Mailchimp }
}
</script>
<style>
.signup-button {
  margin: .75rem 0rem;
}
.benefits {
  margin: 1rem -.7rem;
}
.subscribe-button {
  margin: 1rem 0rem;
}
.footer {
  width: 100%;
}
.benefits {
  margin: 2rem;
}
.screenshot {
  margin: .25rem;
}
.screen {
  border: solid 2px;
  border-radius: 5px;
}
.what-is-edhgo {
  margin: 1rem;
}
</style>
