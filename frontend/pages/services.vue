<template>
  <section class="section">
    <div class="section-header">
      <h2 class="section-title">Service directory</h2>
      <span class="badge">Offline</span>
    </div>
    <div class="empty-state">
      <p>Service telemetry has not been connected. Pulse will show N/A until data is available.</p>
    </div>
  </section>

  <section class="section">
    <table class="table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Protocol</th>
          <th>Entry</th>
          <th>Exit</th>
          <th>Status</th>
          <th>Users</th>
          <th>Traffic</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>N/A</td>
          <td>Unknown</td>
          <td>N/A</td>
          <td>N/A</td>
          <td>Offline</td>
          <td>N/A</td>
          <td>N/A</td>
        </tr>
      </tbody>
    </table>
  </section>
</template>

<script setup lang="ts">
useHead({ title: 'Services' })
</script>
