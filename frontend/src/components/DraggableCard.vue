<template>
  <div ref="myDraggable" class="draggable">
    <!--  Wrap Card here -->
    <Card v-bind="card" :ScreenX="screenX" v-bind:ScreenY="screenY"/>
  </div>
</template>

<script>
import interact from 'interactjs';
import Card from '@/components/Card';
export default {
  name: 'DraggableCard',
  data() {
    return {
      screenX: 0,
      screenY: 0,
      lastSource: '',
      lastDestination: '',
    };
  },
  props: {
    card: Object,
    user: Object,
  },
  mounted: function () {
    let myDraggable = this.$refs.myDraggable;
    this.initInteract(myDraggable);
  },
  methods: {
    initInteract: function (selector) {
      interact(selector).draggable({
        // enable inertial throwing
        inertia: true,
        // enable autoScroll
        autoScroll: true,
        // call this function on every dragmove event
        onmove: this.dragMoveListener,
        // call this function on every dragend event
        onend: this.onDragEnd,
      });
    },
    dragMoveListener: function (event) {
      var target = event.target,
        // keep the dragged position in the data-x/data-y attributes
        x = (parseFloat(target.getAttribute('data-x')) || this.screenX) + event.dx,
        y = (parseFloat(target.getAttribute('data-y')) || this.screenY) + event.dy;

      // translate the element
      target.style.webkitTransform = target.style.transform = 'translate(' + x + 'px, ' + y + 'px)';

      // update the position attributes
      target.setAttribute('data-x', x);
      target.setAttribute('data-y', y);

      if (!!event.dragLeave) {
        this.lastSource = event.dragLeave.id
      }
      if (!!event.dragEnter) {
        this.lastDestination = event.dragEnter.id
        this.card.CurrentZone = this.lastDestination
      }
    },
    onDragEnd: function (event) {
      var target = event.target;
      this.screenX = target.getBoundingClientRect().left;
      this.screenY = target.getBoundingClientRect().top;
      this.card.ScreenX = this.screenX
      this.card.ScreenY = this.screenY
      this.move({ 
        source: this.lastSource, 
        destination: this.lastDestination, 
        card: this.card
      })
      let g = this.$store.state.Games.game
      this.$store.dispatch('Games/sync', g)
    },
    capitalize: function (word) {
      return word[0].toUpperCase() + word.slice(1).toLowerCase();
    },
    move: function({ source, destination, card}) {
      console.log('source: ', source)
      console.log('destination: ', destination)
      console.log('card: ', card)
    },
    checkOverlap(elem1, elem2) {
      const rect1 = elem1.getBoundingClientRect();
      const rect2 = elem2.getBoundingClientRect();
      return !(rect2.left > rect1.right ||
                rect2.right < rect1.left ||
                rect2.top > rect1.bottom ||
                rect2.bottom < rect1.top);
    },
  },
  components: {
    Card,
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.draggable {
  padding: 5px;
  position: absolute;
}
</style>
