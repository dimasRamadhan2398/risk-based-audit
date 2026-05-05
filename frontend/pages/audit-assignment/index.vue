<template>
  <div class="grid grid-cols-2 gap-4">
    <div>
      <label class="block text-sm font-medium text-gray-700"
        >Activity ID</label
      >
      <input
        v-model="form.activityId"
        required
        type="text"
        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm p-2 border"
        placeholder="e.g. AUD-001"
      />
    </div>
    <div>
      <label class="block text-sm font-medium text-gray-700"
        >Letter Number</label
      >
      <input
        v-model="form.letterNumber"
        required
        type="text"
        class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm p-2 border"
        placeholder="e.g. AUD-001"
      />
    </div>
  </div>
</template>
<script>
export default {
  data() {
    return {
      form: {
        activityId: "",
        letterNumber: "",
      },
    };
  },
};
</script>
