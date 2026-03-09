<template>
    <Teleport to="body">
      <div v-if="store.showModalF01" class="fixed inset-0 bg-gray-900/60 flex items-center justify-center p-4">
        <div class="bg-secondary-50 dark:bg-secondary-300 rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <UForm @submit.prevent="store.handleSubmitF01">
        
        <div class="px-6 py-4 border-b border-secondary-200 dark:border-secondary-700 bg-secondary-50 dark:bg-secondary-900 rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 dark:text-white">Assignment Reference</h3>
            <UIcon name="close" @click="store.closeModalF01" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>

        <div class="space-y-6 m-6">
        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Assignment Letter<span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.headerForm.assignmentLetterId" :items="store.options.assignmentLetter" placeholder="Choose Assignment Letter" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Audit Purpose<span class="text-red-500">*</span></label>
            <UInput class="md:col-span-3" v-model="store.headerForm.auditPurpose" disabled placeholder="(Automatically filled in when filling out the assignment letter)" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Business Process<span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.headerForm.businessProcess" :items="store.options.businessProcess" placeholder="Choose Business Process" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300">Audit Period<span class="text-red-500">*</span></label>
            <div class="md:col-span-3">
            <UFormField :error="store.dateErrorMessage">
                <div class="flex items-center gap-4 w-full">
                <UInput 
                    type="date" 
                    v-model="store.headerForm.periodStart" 
                    icon="i-heroicons-calendar" 
                    class="w-full"
                    :color="store.isDateError ? 'error' : 'neutral'"
                />
                
                <span class="text-gray-500 font-bold whitespace-nowrap">s/d</span>
                
                <UInput 
                    type="date" 
                    v-model="store.headerForm.periodEnd" 
                    icon="i-heroicons-calendar" 
                    class="w-full"
                    :color="store.isDateError ? 'error' : 'neutral'"
                />
                </div>
            </UFormField>
            </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Location<span class="text-red-500">*</span></label>
            <USelectMenu class="md:col-span-3" v-model="store.headerForm.location" :items="store.options.location" placeholder="Choose Location" />
        </div>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <label class="font-semibold text-sm text-gray-700 dark:text-gray-300 mt-2">Team</label>
            <div class="md:col-span-3 space-y-4">

            <div v-for="(member, index) in store.headerForm.teamMembers" :key="member.id" class="flex gap-2 items-center">
                <div class="grid grid-cols-2 gap-2 flex-1">
                <USelectMenu 
                    v-model="member.name" 
                    :items="store.getAvailableMembers(index)" 
                    placeholder="Choose Member"
                />
                <UInput 
                    v-model="member.role" 
                    placeholder="Position" 
                />
                </div>

                <UButton 
                v-if="store.headerForm.teamMembers.length > 1"
                icon="i-heroicons-trash" 
                color="error" 
                variant="ghost" 
                @click="store.removeTeamMember(index)" 
                />
            </div>

            <UButton 
                color="primary" 
                variant="soft"
                icon="i-heroicons-plus" 
                label="Add Member" 
                @click="store.addTeamMember()"
            />
            </div>
        </div>
        

        <div class="flex justify-end pt-10 border-gray-100 dark:border-gray-800">
            <UButton 
                :label="store.isEditingF01 ? 'Update Data' : 'Submit'" 
                color="primary"
                @click="store.handleSubmitF01" 
            />
        </div>
        </div>
    </UForm>
    </div>
    </div>
    </Teleport>
</template>

<script setup lang="ts">
import { useWorkingPaperStore } from '~/stores/working-paper'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useWorkingPaperStore()
</script>