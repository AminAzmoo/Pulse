<script setup lang="ts">
useHead({ title: 'Pulse | Services' });
</script>

<template>
  <section class="page-header">
    <h1 class="page-title">Services</h1>
    <p class="page-subtitle">
      Service health and dependency visibility.
    </p>
  </section>

  <section class="section">
    <h2 class="section-title">Service Health</h2>
    <table class="table">
      <thead>
        <tr>
          <th>Service</th>
          <th>Status</th>
          <th>Latency</th>
          <th>Owner</th>
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
    <p class="empty-state">No service telemetry has been ingested yet.</p>
  </section>
</template>
