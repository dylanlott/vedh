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
    };
  },
  props: {
    card: Object,
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

      // TODO: These are the events we need to access in order to figure out where to put cards.
      // console.log('dragEnd: ', event.dragEnd)
      
      console.log(this.card.ID, ' is leaving ', event.dragLeave)
      // TODO: we have everything we need from this.card  and the destination to write our logic. 
      // This is the best way to handle this I think, with a card-centric view.
    },
    onDragEnd: function (event) {
      var target = event.target;
      // update the state
      this.screenX = target.getBoundingClientRect().left;
      this.screenY = target.getBoundingClientRect().top;

      // TODO: update the server's boardstate view

      // console.log('card dropped: ', this.card)
      // console.log('dropped event: ', event)
      // console.log('caller: ', event.target || event.srcElement)
      // console.log('dropped into: ', event.dropzone.target || undefined)
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
