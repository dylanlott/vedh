<template>
  <div ref="myDraggable" class="draggable">
    <!--  Wrap Card here -->
    <Card v-bind="card" />
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

      // update the posiion attributes
      target.setAttribute('data-x', x);
      target.setAttribute('data-y', y);

      // save last known source and destination events
      if (!!event.dragLeave) {
        this.lastSource = event.dragLeave.id
      }

      if (!!event.dragEnter) {
        this.lastDestination = event.dragEnter.id
      }

      if (this.lastDestination && this.lastSource) {
        // TODO: call move action and reset lastSource and lastDestination fields
      }
    },
    onDragEnd: function (event) {
      var target = event.target;
      // update the state

      // TODO: do we need to persist this into the server as well?
      this.screenX = target.getBoundingClientRect().left;
      this.screenY = target.getBoundingClientRect().top;
    },
    capitalize: function (word) {
      return word[0].toUpperCase() + word.slice(1).toLowerCase();
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
