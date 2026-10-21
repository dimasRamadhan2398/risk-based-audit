<template>
  <div class="p-4 sm:p-6 min-h-screen overflow-x-hidden">
    <!-- 1. Auditor Dashboard (Role: Auditor, Audit Staff) -->
    <div v-if="currentDashboardRole === 'auditor'" key="auditor-dashboard-container">
      <AuditorDashboard />
    </div>

    <!-- 2. Auditee Dashboard (Role: Auditee, Department Head / PIC Unit) -->
    <div v-else-if="currentDashboardRole === 'auditee'" key="auditee-dashboard-container">
      <AuditeeDashboard />
    </div>

    <!-- 3. Executive Dashboard (Role: Chief Audit Executive, Viewer, Director) -->
    <div v-else-if="currentDashboardRole === 'executive'" key="executive-dashboard-container">
      <ExecutiveDashboard />
    </div>

    <!-- 4. Admin Dashboard (Role: Admin / Full Macro View) -->
    <div v-else key="admin-dashboard-container">
      <AdminDashboard />
    </div>
  </div>
</template>

<script setup lang="ts">
import AuditorDashboard from "~/components/dashboard/AuditorDashboard.vue";
import AuditeeDashboard from "~/components/dashboard/AuditeeDashboard.vue";
import ExecutiveDashboard from "~/components/dashboard/ExecutiveDashboard.vue";
import AdminDashboard from "~/components/dashboard/AdminDashboard.vue";
import { useRbac } from "~/composables/useRbac";

definePageMeta({
  middleware: "auth",
});

const { currentDashboardRole } = useRbac();
</script>
