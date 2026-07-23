<template>
    
      <UModal v-model:open="store.isModalOpen" :dismissible="false" :ui="{ content: 'sm:max-w-2xl bg-[var(--bg-main)] border border-[var(--border-main)]' }">
        <div></div>
        <template #content>
        <UForm 
        :state="store.form"
        @submit.prevent="store.handleSubmit"
        >
        <div class="relative  rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
          <div class="flex justify-between items-center p-6 border-b border-gray-100">
            <h2 class="text-xl font-bold text-gray-800  flex items-center gap-2">
              <UIcon name="i-heroicons-document-plus" class="w-6 h-6 text-orange-500" />
              Add Assignment Letter
            </h2>
            <UIcon name="i-heroicons-x-mark" @click="store.closeModal" size="xl" />
          </div>

          <div class="p-8 overflow-y-auto space-y-6 flex-1">

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 ">Audit Title</label>
              <UFormField class="md:col-span-3">
                <UInput v-model="store.form.auditTitle" placeholder="Example: Financial Operational Audit" size="lg" class="w-full"/>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 ">Led By (PIC)</label>
              <UFormField class="md:col-span-3">
                <UInput v-model="store.form.leader" placeholder="Person in Charge Name" size="lg" class="w-full"/>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 ">Audit Category</label>
              <UFormField class="md:col-span-3">
                <USelectMenu v-model="store.form.category" :items="categoryOptions" size="lg" class="w-full" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 ">Audit Year</label>
              <UFormField class="md:col-span-3">
                <div class="flex items-center gap-4">
                  <UInput v-model="store.form.auditYear" type="date" size="lg" class="flex-1 w-full" />
                </div>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center border-y border-gray-100  py-4 my-6">
            <label class="font-bold text-gray-700 ">Audit Team</label>
            <UFormField class="md:col-span-3">
                <URadioGroup
                orientation="horizontal"
                variant="list"
                default-value="System"
                :items="store.options.auditTeam"
                />
            </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700  mt-2">Audit Period</label>
              <UFormField class="md:col-span-3">
                <div class="flex items-center gap-4">
                  <UInput v-model="store.form.startPeriod" type="date" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                  <span class="font-bold text-gray-500">to</span>
                  <UInput v-model="store.form.finishPeriod" type="date" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                </div>
                <p v-if="store.dateError" class="text-red-500 text-sm font-semibold mt-1">{{ store.dateError }}</p>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 ">Work Unit</label>
              <UFormField class="md:col-span-3">
                <USelectMenu v-model="store.form.workingUnit" :items="store.options.workingUnit" placeholder="Select Work Unit (Required)" size="lg" class="w-full" :popper="{ strategy: 'absolute' }"/>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700  mt-2">Team Members</label>
              <UFormField class="md:col-span-3 space-y-3">
                <div v-for="(member, index) in store.form.membersList" :key="index" class="flex items-center gap-2">
                  <UInput v-model="member.name" placeholder="Member Name" class="flex-1" />
                  <USelectMenu v-model="member.role" :items="store.options.role" placeholder="Role" class="w-1/3" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.membersList, index)" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add Member" @click="store.addItem(store.form.membersList, { name: '', role: 'Member' })" />
                <p class="text-md text-orange-600  font-semibold">* According to template, minimum 3 team members is recommended.</p>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700  mt-2">Audit Purpose</label>
              <UFormField class="md:col-span-3 space-y-3">
                <div v-for="(purpose, index) in store.form.purposeList" :key="index" class="flex items-start gap-2">
                  <span class="mt-2 font-bold text-gray-400">{{ index + 1 }}.</span>
                  <UTextarea v-model="store.form.purposeList[index]" placeholder="Type audit purpose..." class="flex-1" :rows="2" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.purposeList, index)" class="mt-1" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add Purpose" @click="store.addItem(store.form.purposeList, '')" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700  mt-2">Scope</label>
              <UFormField class="md:col-span-3 space-y-3">
                <div v-for="(scope, index) in store.form.scopeList" :key="index" class="flex items-start gap-2">
                  <UInput v-model="store.form.scopeList[index]" placeholder="Type scope..." class="flex-1" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.scopeList, index)" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add Scope" @click="store.addItem(store.form.scopeList, '')" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700  mt-2">CC (Carbon Copy)</label>
              <UFormField class="md:col-span-3 space-y-3">
                <div v-for="(cc, index) in store.form.ccList" :key="index" class="flex items-center gap-2">
                  <UInput v-model="store.form.ccList[index]" placeholder="CC Position (Example: President Director)" class="flex-1" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.ccList, index)" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add CC" @click="store.addItem(store.form.ccList, '')" />
              </UFormField>
            </div>

          </div>

          <div class="p-6 border-t border-gray-100 flex justify-end items-center gap-4">
            <!-- <UButton label="Cancel" color="neutral" variant="soft" @click="store.closeModal" class="font-bold text-gray-500 hover:text-gray-700 px-4 py-2" /> -->
            <UButton label="Save Assignment Letter" color="primary" size="lg" class="font-bold px-8 shadow-md" @click="store.handleSubmit" />
          </div>
        </div>
        </UForm>
        </template>
      </UModal>
      
</template>

<script setup lang="ts">
import { useAssignmentLetterStore } from '~/stores/assignment-letter'

const store = useAssignmentLetterStore()

const categoryOptions = ['Assurance', 'Special Audit', 'Specific Reason', 'Consulting Services', 'Follow-Up Audit', 'Investigation', 'Quality Assurance Review']

</script>
