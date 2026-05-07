<template>
    <UModal>
        <template #content>
            <UForm>
                <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
                    <UIcon name="charter" class="text-primary-500 " size="32"></UIcon>
                    <h3 class="text-lg font-bold text-[var(--text-main)]">Audit Result Report Form</h3>
                    <UIcon name="close" @click="store.closeModal" class="text-primary-400 hover:text-primary-600  text-2xl cursor-pointer"></UIcon>
                </div>

                <div class="space-y-4">
                    <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">Basic Information</h4>
                    
                    
                </div>
            </UForm>
        </template>
    </UModal>
</template>

<script>

</script>