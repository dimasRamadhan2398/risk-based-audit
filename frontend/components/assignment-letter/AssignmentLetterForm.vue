<template>
      <UModal 
        v-model:open="store.isModalOpen" 
        :dismissible="false" 
        :ui="{
            content: 'sm:max-w-4xl max-h-[90vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
            header: 'border-b border-gray-100 dark:border-gray-800 p-5 text-gray-900 dark:text-white font-bold shrink-0',
            body: 'p-6 space-y-5 bg-white dark:bg-gray-900 text-gray-900 dark:text-white overflow-y-auto max-h-[calc(90vh-130px)] flex-1',
            footer: 'border-t border-gray-100 dark:border-gray-800 p-4 shrink-0 bg-white dark:bg-gray-900',
            overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
        }"
      >
        <div></div>
        <template #content>
        <UForm 
        :state="store.form"
        @submit.prevent="store.handleSubmit"
        >
        <div class="relative  rounded-xl shadow-2xl flex flex-col max-h-[90vh]">
          <div class="flex justify-between items-center p-6 border-b border-gray-100">
            <h2 class="text-xl font-bold text-gray-800  flex items-center gap-2">
              <UIcon :name="store.editingId ? 'i-heroicons-pencil-square' : 'i-heroicons-document-plus'" class="w-6 h-6 text-orange-500" />
              {{ store.editingId ? 'Edit Assignment Letter' : 'Add Assignment Letter' }}
            </h2>
            <UIcon name="i-heroicons-x-mark" @click="store.closeModal" size="xl" />
          </div>

          <div class="p-8 overflow-y-auto space-y-6 flex-1">

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <UFormField label="Audit Title" class="md:col-span-4" required>
                <UInput 
                  v-model="store.form.auditTitle" 
                  placeholder="Example: Financial Operational Audit" 
                  size="lg" 
                  class="w-full"
                  required
                  maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity('Judul audit maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                />
                <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.form.auditTitle ? store.form.auditTitle.length : 0 }}/100
                </div>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <UFormField label="Audit Category" class="md:col-span-4" required>
                <USelectMenu v-model="store.form.category" :items="categoryOptions" size="lg" class="w-full" required/>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <UFormField label="Audit Year" class="md:col-span-4" required>
                  <UInput v-model="store.form.auditYear" type="date" size="lg" class="flex-1 w-full" required/>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center border-y border-gray-100  py-4 my-6">
              <label class="font-bold text-gray-700 ">Audit Team</label>
              <UFormField class="md:col-span-3">
                <div class="flex flex-col gap-2">
                  <label v-for="team in store.options.auditTeam" :key="team.value" class="inline-flex items-center gap-2 cursor-pointer">
                    <input
                      type="radio"
                      v-model="store.form.auditTeam"
                      :value="team.value"
                      class="accent-orange-500 w-4 h-4"
                      required
                    />
                    <span class="text-sm font-medium text-gray-800 dark:text-white">{{ team.label }}</span>
                  </label>
                </div>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <UFormField label="Audit Period" class="md:col-span-4" required>
                <div class="flex items-center gap-4">
                  <AppDatePicker v-model="store.form.startPeriod" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                  <span class="font-bold text-gray-500">to</span>
                  <AppDatePicker v-model="store.form.finishPeriod" size="lg" class="flex-1" :class="{'ring-red-500': store.dateError}" />
                </div>
                <p v-if="store.dateError" class="text-red-500 text-sm font-semibold mt-1">{{ store.dateError }}</p>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-center">
              <UFormField label="Work Unit" class="md:col-span-4" required>
                <USelectMenu v-model="store.form.workingUnit" :items="store.options.workingUnit" placeholder="Select Work Unit (Required)" size="lg" class="w-full" :popper="{ strategy: 'absolute' }" required/>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <UFormField label="Team Members" class="md:col-span-4 space-y-3" required>
                <div v-for="(member, index) in store.form.membersList" :key="index" class="flex items-center gap-2">
                  <UInput 
                  v-model="member.name" 
                  placeholder="Member Name" 
                  class="flex-1" 
                  required
                  maxlength="100"
                  @invalid="($event.target as any)?.setCustomValidity('Nama anggota maksimal 100 karakter dan wajib diisi')"
                  @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ member.name ? member.name.length : 0 }}/100
                  </div>
                  <USelectMenu v-model="member.role" :items="store.options.role" placeholder="Role" class="w-1/3" required/>
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.membersList, index)" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add Member" @click="store.addItem(store.form.membersList, { name: '', role: 'Member' })" />
                <p class="text-md text-orange-600  font-semibold">* According to template, minimum 3 team members is recommended.</p>
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <UFormField label="Audit Purpose" class="md:col-span-4 space-y-3">
                <div v-for="(purpose, index) in store.form.purposeList" :key="index" class="flex items-start gap-2">
                  <span class="mt-2 font-bold text-gray-400">{{ index + 1 }}.</span>
                  <UTextarea v-model="store.form.purposeList[index]" placeholder="Type audit purpose..." class="flex-1" :rows="2"/>
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.purposeList, index)" class="mt-1" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add Purpose" @click="store.addItem(store.form.purposeList, '')" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <UFormField label="Scope" class="md:col-span-4 space-y-3">
                <div v-for="(scope, index) in store.form.scopeList" :key="index" class="flex items-start gap-2">
                  <UInput v-model="store.form.scopeList[index]" placeholder="Type scope..." class="flex-1"/>
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.scopeList, index)" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add Scope" @click="store.addItem(store.form.scopeList, '')" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-4 gap-4 items-start">
              <UFormField label="CC (Carbon Copy)" class="md:col-span-4 space-y-3">
                <div v-for="(cc, index) in store.form.ccList" :key="index" class="flex items-center gap-2">
                  <UInput 
                    v-model="store.form.ccList[index]" 
                    placeholder="CC Position (Example: President Director)" 
                    class="flex-1"
                    required
                    maxlength="100"
                    @invalid="($event.target as any)?.setCustomValidity('CC Position maksimal 100 karakter dan wajib diisi')"
                    @input="($event.target as any)?.setCustomValidity('')"
                  />
                  <div class="text-xs text-gray-500 mt-1 text-right">
                    {{ store.form.ccList[index] ? store.form.ccList[index].length : 0 }}/100
                  </div>
                  <UButton icon="i-heroicons-trash" color="error" variant="ghost" @click="store.removeItem(store.form.ccList, index)" />
                </div>
                <UButton class="mt-4" color="primary" variant="soft" icon="i-heroicons-plus" label="Add CC" @click="store.addItem(store.form.ccList, '')" />
              </UFormField>
            </div>
          </div>

          <!-- Error message display -->
          <div v-if="store.errorMsg" class="m-6 mt-0 p-3 bg-error-50 text-error-700 rounded-lg text-sm font-semibold">
            {{ store.errorMsg }}
          </div>

          <div class="p-6 border-t border-gray-100 flex justify-end items-center gap-4">
            <UButton :label="store.editingId ? 'Update Assignment Letter' : 'Save Assignment Letter'" color="primary" size="lg" class="font-bold px-8 shadow-md" @click="store.handleSubmit"/>
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
