<template>
    <UModal
        title="Add Risk Mitigation"
        :close="{
            color: 'primary',
            variant: 'outline',
            class: 'rounded-full'
        }"
    >
        <template #body>
            <UForm>
                <UFormField
                    orientation="horizontal"
                    label="Risk Mitigation Name"
                    help="Masukkan kejadian yang ingin ditinjau yang berpotensi menimbulkan dampak negatif terhadap tujuan, proses, atau aset."
                    class="text-sm font-medium"
                    size="lg"
                    :ui="{
                        root: 'wrap-break-word flex flex-row flex-1 w-full basis-0.5',
                        container: ' gap-4 items-start max-w-full',
                        label: 'block font-medium text-black min-w-24 max-w-24',
                    }"
                >
                    <UInput></UInput>
                </UFormField>
            </UForm>
        </template>
        <template #footer>
            <div></div>
            <UButton label="Confirm" color="primary"></UButton>
        </template>
    </UModal>
</template>

<script setup lang="ts">

</script>