<script setup lang="ts">
useHead({ title: 'Pulse | Devices' });
</script>

<template>
  <section class="page-header">
    <h1 class="page-title">Devices</h1>
    <p class="page-subtitle">
      Inventory of connected edge devices and gateways.
    </p>
  </section>

  <section class="section">
    <h2 class="section-title">Device Inventory</h2>
    <table class="table">
      <thead>
        <tr>
          <th>Device</th>
          <th>Status</th>
          <th>Last Seen</th>
          <th>Location</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>N/A</td>
          <td>Unknown</td>
          <td>N/A</td>
          <td>N/A</td>
        </tr>
      </tbody>
    </table>
    <p class="empty-state">Connect the Pulse API to populate device data.</p>
  </section>
</template>
