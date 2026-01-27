<template id="edhgo">
  <div>
    <Navbar></Navbar>
      <router-view/>
  </div>
</template>

<script>
import Navbar from '@/components/Navbar.vue'
export default {
  components: {
    Navbar
  }
};
</script>