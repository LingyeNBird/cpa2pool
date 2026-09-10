<script setup lang="ts" generic="T">
import { onBeforeUnmount, onBeforeUpdate, onUpdated, ref } from 'vue';
import './ReportTable.css';
const props = defineProps<{
  items: T[];
  rowKey: (item: T) => string;
  columns: number;
  empty: string;
  pending?: boolean;
}>();
const root = ref<HTMLElement>();
let previousHeight = 0;
let resize: Animation | undefined;
let columnsPinned = false;
const bounds = new Map<Element, { row: DOMRect; cells: number[] }>();
const departures = new Set<{
  element: Element;
  ghost: HTMLElement;
  animation: Animation;
  done: () => void;
}>();
const reduced = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;
onBeforeUpdate(() => {
  const element = root.value;
  if (!element) return;
  previousHeight = element.getBoundingClientRect().height;
  bounds.clear();
  for (const row of element.querySelectorAll(
    ':scope > table:not(.report-row-ghost) > tbody > tr',
  )) {
    bounds.set(row, {
      row: row.getBoundingClientRect(),
      cells: Array.from(row.children, (cell) => cell.getBoundingClientRect().width),
    });
  }
});
onUpdated(() => {
  const element = root.value;
  if (!element) return;
  resize?.cancel();
  const height = element.getBoundingClientRect().height;
  if (!columnsPinned && props.items.length) {
    const table = element.querySelector('table')!;
    const headers = Array.from(table.querySelectorAll('thead th')) as HTMLElement[];
    const width = table.getBoundingClientRect().width;
    const widths = headers.map((header) => header.getBoundingClientRect().width);
    headers.forEach((header, index) => {
      header.style.width = `${(widths[index] / width) * 100}%`;
    });
    table.style.tableLayout = 'fixed';
    columnsPinned = true;
  }
  if (!reduced() && previousHeight && Math.abs(height - previousHeight) > 0.5) {
    resize = element.animate([{ height: `${previousHeight}px` }, { height: `${height}px` }], {
      duration: 220,
      easing: 'cubic-bezier(0.23, 1, 0.32, 1)',
    });
  }
});
function leave(element: Element, done: () => void) {
  const container = root.value;
  const snapshot = bounds.get(element);
  if (!container || !snapshot || reduced()) {
    done();
    return;
  }
  const parent = container.getBoundingClientRect();
  const ghost = document.createElement('table');
  ghost.className = 'table report-row-ghost';
  ghost.setAttribute('aria-hidden', 'true');
  ghost.inert = true;
  const copy = element.cloneNode(true) as HTMLElement;
  copy.removeAttribute('id');
  copy.querySelectorAll('[id]').forEach((node) => node.removeAttribute('id'));
  copy.style.transform = 'none';
  copy.style.transition = 'none';
  Array.from(copy.children).forEach((cell, index) => {
    (cell as HTMLElement).style.width = `${snapshot.cells[index]}px`;
    (cell as HTMLElement).style.borderBottom = getComputedStyle(
      element.children[index],
    ).borderBottom;
  });
  copy.style.backgroundColor = getComputedStyle(element).backgroundColor;
  const body = document.createElement('tbody');
  body.append(copy);
  ghost.append(body);
  ghost.style.left = `${snapshot.row.left - parent.left - container.clientLeft + container.scrollLeft}px`;
  ghost.style.top = `${snapshot.row.top - parent.top - container.clientTop + container.scrollTop}px`;
  ghost.style.width = `${snapshot.row.width}px`;
  container.append(ghost);
  (element as HTMLElement).style.display = 'none';
  const animation = ghost.animate(
    [
      { opacity: 1, transform: 'translateX(0)' },
      { opacity: 0, transform: 'translateX(-8px)' },
    ],
    { duration: 180, easing: 'cubic-bezier(0.23, 1, 0.32, 1)' },
  );
  const departure = { element, ghost, animation, done };
  departures.add(departure);
  const finish = () => {
    ghost.remove();
    departures.delete(departure);
    done();
  };
  animation.onfinish = finish;
  animation.oncancel = finish;
}
function cancelLeave(element: Element) {
  (element as HTMLElement).style.removeProperty('display');
  for (const departure of departures) {
    if (departure.element === element) departure.animation.cancel();
  }
}
onBeforeUnmount(() => {
  resize?.cancel();
  for (const departure of departures) departure.animation.cancel();
});
</script>
<template>
  <div
    ref="root"
    class="table-wrap table-list table-list-inset report-table"
    :aria-busy="pending || undefined"
  >
    <table class="table">
      <thead>
        <tr>
          <slot name="header" />
        </tr>
      </thead>
      <TransitionGroup tag="tbody" name="report-row" @leave="leave" @leave-cancelled="cancelLeave">
        <tr v-for="item in items" :key="rowKey(item)">
          <slot name="row" :item="item" />
        </tr>
        <tr v-if="!items.length" key="empty-row">
          <td :colspan="columns" class="report-table-empty">{{ pending ? '\u00a0' : empty }}</td>
        </tr>
      </TransitionGroup>
    </table>
  </div>
</template>
