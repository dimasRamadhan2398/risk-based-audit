<template>
    
      <UModal 
        v-model:open="store.showModalF01" 
        :dismissible="false" 
        :ui="{
            content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
            body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
            footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
        }"
      >
        <template #content>
        <UForm :schema="headerSchema" :state="store.headerForm" @submit.prevent="store.handleSubmitF01">
        <div class="rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-y-auto">
        <div class="px-6 py-4 border-b border-secondary-200  rounded-t-xl flex justify-between items-center">
            <UIcon name="charter" class=" text-primary-500" size="32"></UIcon>
            <h3 class="text-lg font-bold text-secondary-900 ">Assignment Reference</h3>
            <UIcon name="close" @click="store.closeModalF01" class="text-primary-400 hover:text-primary-600 text-2xl">&times;</UIcon>
        </div>

        <div class="space-y-6 m-6">
        <UFormField 
            label="Assignment Letter" 
            name="assignmentLetterId" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <USelectMenu v-model="store.headerForm.assignmentLetterId" :items="store.options.assignmentLetter" placeholder="Choose Assignment Letter" class="w-full" />
        </UFormField>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Audit Purpose" class="font-semibold text-sm text-gray-700  mt-2" />
            <UInput class="md:col-span-3" v-model="store.headerForm.auditPurpose" disabled placeholder="(Automatically filled in when filling out the assignment letter)" />
        </div>

        <UFormField 
            label="Business Process Name" 
            name="businessProcess" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <UInput 
                v-model="store.headerForm.businessProcess" 
                class="w-full" 
                :maxlength="100"
                @invalid="($event.target as any)?.setCustomValidity('Business Process Name maksimal 100 karakter dan wajib diisi')"
                @input="($event.target as any)?.setCustomValidity('')"
            />
            <div class="text-xs text-gray-500 mt-1 text-right">
                {{ store.headerForm.businessProcess ? store.headerForm.businessProcess.length : 0 }}/100
            </div>
        </UFormField>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Activities" class="font-semibold text-sm text-gray-700 mt-2 w-full" />
            <div class="md:col-span-3 space-y-4 ">
                <div v-for="(activity, index) in store.headerForm.activities" :key="activity.id" class="flex gap-2 items-start">
                    <div class="flex-1">
                        <UTextarea 
                            v-model="activity.name" 
                            placeholder="Activity Description"
                            autoresize
                            :rows="2"
                            :maxrows="5"
                            class="w-full"
                        />
                    </div>
                    <UButton 
                        v-if="store.headerForm.activities.length > 1"
                        icon="i-heroicons-trash" 
                        color="error" 
                        variant="ghost" 
                        class="mt-1"
                        @click="store.removeActivity(index)" 
                    />
                </div>
                <UButton 
                    color="primary" 
                    variant="soft"
                    icon="i-heroicons-plus" 
                    label="Add Activity" 
                    @click="store.addActivity()"
                />
            </div>
        </div>

        <UFormField 
            label="Audit Period" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <div class="flex items-center gap-4 w-full">
                <UFormField name="periodStart" class="w-full">
                    <AppDatePicker 
                        v-model="store.headerForm.periodStart" 
                        class="w-full"
                    />
                </UFormField>
                <span class="text-gray-500 font-bold whitespace-nowrap">s/d</span>
                <UFormField name="periodEnd" class="w-full">
                    <AppDatePicker 
                        v-model="store.headerForm.periodEnd" 
                        class="w-full"
                    />
                </UFormField>
            </div>
        </UFormField>

        <UFormField 
            label="Location" 
            name="location" 
            required 
            class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10" 
            :ui="{ container: 'md:col-span-3 w-full', label: 'font-semibold text-sm text-gray-700 mt-2' }"
        >
            <USelectMenu v-model="store.headerForm.location" :items="store.options.location" placeholder="Choose Location" class="w-full" />
        </UFormField>

        <div class="grid grid-cols-1 md:grid-cols-4 items-start max-w-full mt-10">
            <UFormField label="Team" class="font-semibold text-sm text-gray-700  mt-2" />
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
        

        <div class="flex justify-end p-6 border-gray-100 ">
            <UButton 
                type="submit"
                :label="store.isEditingF01 ? 'Update Data' : 'Submit'" 
                color="primary"
            />
        </div>
        </div>
    </div>
    </UForm>
    </template>
    </UModal>
    
</template>

<script setup lang="ts">
import { useWorkingPaperStore, headerSchema } from '~/stores/working-paper'

const store = useWorkingPaperStore()
</script>