<template>
    <Teleport to="body">
      <div v-if="store.isModalOpen" class="fixed inset-0 bg-gray-900/60 flex items-center justify-center p-4">
        <div class="bg-white dark:bg-gray-900 rounded-xl shadow-2xl w-full max-w-5xl max-h-[95vh] flex flex-col overflow-hidden">

          <div class="flex justify-between items-center p-6 border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
            <h2 class="text-xl font-bold text-gray-800 dark:text-white flex items-center gap-2">
              <UIcon name="i-heroicons-document-plus" class="w-6 h-6 text-orange-500" />
              Add Assignment Letter
            </h2>
            <button @click="store.closeModal" class="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300">
              <UIcon name="i-heroicons-x-mark" class="w-7 h-7" />
            </button>
          </div>

          <div class="p-8 overflow-y-auto space-y-6 flex-1">

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Audit Title</label>
              <div class="md:col-span-3">
                <UInput v-model="store.form.auditTitle" placeholder="Example: Financial Operational Audit" size="lg" class="w-full"/>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Led By (PIC)</label>
              <div class="md:col-span-3">
                <UInput v-model="store.form.leader" placeholder="Person in Charge Name" size="lg" class="w-full"/>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Audit Category</label>
              <div class="md:col-span-3">
                <UFormField size="lg">
                    <select v-model="store.form.category" class="input-field bg-white rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500">
                        <option value="Assurance">Assurance</option>
                        <option value="Special Audit">Special Audit</option>
                        <option value="Specific Reason">Specific Reason</option>
                        <option value="Consulting Services">Consulting Services</option>
                        <option value="Investigation">Investigation</option>
                        <option value="Quality Assurance Review">Quality Assurance Review</option>
                        <option value="Follow-Up Audit">Follow-Up Audit</option>
                    </select>
                </UFormField>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Audit Year</label>
              <div class="md:col-span-3">
                <div class="flex items-center gap-4">
                  <UInput v-model="store.form.auditYear" type="date" size="lg" class="flex-1 w-full" />
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center border-y border-gray-100 dark:border-gray-800 py-4 my-6">
            <label class="font-bold text-gray-700 dark:text-gray-300">Audit Team</label>
            <div class="md:col-span-3">
                <URadioGroup
                orientation="horizontal"
                variant="list"
                default-value="System"
                :items="store.options.auditTeam"
                />
            </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Audit Period</label>
              <div class="md:col-span-3">
                <div class="flex items-center gap-4">
                  <UInput v-model="store.form.startPeriod" type="date" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                  <span class="font-bold text-gray-500">to</span>
                  <UInput v-model="store.form.finishPeriod" type="date" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                </div>
                <p v-if="store.dateError" class="text-red-500 text-sm font-semibold mt-1">{{ store.dateError }}</p>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <label class="font-bold text-gray-700 dark:text-gray-300">Work Unit</label>
              <div class="md:col-span-3">
                <USelectMenu v-model="store.form.workingUnit" :items="store.options.workingUnit" placeholder="Select Work Unit (Required)" size="lg" class="w-full" :popper="{ strategy: 'absolute' }"/>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Team Members</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(member, index) in store.form.membersList" :key="index" class="flex items-center gap-2">
                  <UInput v-model="member.name" placeholder="Member Name" class="flex-1" />
                  <USelectMenu v-model="member.role" :items="store.options.role" placeholder="Role" class="w-1/3" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.membersList, index)" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Add Member" @click="store.addItem(store.form.membersList, { name: '', role: 'Member' })" />
                <p class="text-xs text-orange-600 dark:text-orange-400 font-semibold">* According to template, minimum 3 team members is recommended.</p>
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Audit Purpose</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(purpose, index) in store.form.purposeList" :key="index" class="flex items-start gap-2">
                  <span class="mt-2 font-bold text-gray-400">{{ index + 1 }}.</span>
                  <UTextarea v-model="store.form.purposeList[index]" placeholder="Type audit purpose..." class="flex-1" :rows="2" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.purposeList, index)" class="mt-1" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Add Purpose" @click="store.addItem(store.form.purposeList, '')" />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">Scope</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(scope, index) in store.form.scopeList" :key="index" class="flex items-start gap-2">
                  <UInput v-model="store.form.scopeList[index]" placeholder="Type scope..." class="flex-1" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.scopeList, index)" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Add Scope" @click="store.addItem(store.form.scopeList, '')" />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <label class="font-bold text-gray-700 dark:text-gray-300 mt-2">CC (Carbon Copy)</label>
              <div class="md:col-span-3 space-y-3">
                <div v-for="(cc, index) in store.form.ccList" :key="index" class="flex items-center gap-2">
                  <UInput v-model="store.form.ccList[index]" placeholder="CC Position (Example: President Director)" class="flex-1" />
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.ccList, index)" />
                </div>
                <UButton color="primary" variant="soft" icon="i-heroicons-plus" label="Add CC" @click="store.addItem(store.form.ccList, '')" />
              </div>
            </div>

          </div>

          <div class="p-6 border-t border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 flex justify-end items-center gap-4">
            <button @click="store.closeModal" class="font-bold text-gray-500 hover:text-gray-700 px-4 py-2">
              Cancel
            </button>
            <UButton
                label="Save Assignment Letter" color="primary" size="lg" class="font-bold px-8 shadow-md" @click="store.handleSubmit" />
          </div>

        </div>
      </div>
    </Teleport>
</template>

<script setup lang="ts">
import { useAssignmentLetterStore } from '~/stores/assignment-letter'
import { AuditCategory } from '~/types/audit';

const store = useAssignmentLetterStore()

</script>
