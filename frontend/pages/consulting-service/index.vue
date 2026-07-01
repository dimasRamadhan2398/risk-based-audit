<template>
  <div class="p-6 space-y-8 max-w-7xl mx-auto min-h-screen">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="space-y-1">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white uppercase tracking-tight">
          Consulting Services Dashboard
        </h1>
        <p class="text-lg text-gray-600 dark:text-gray-400 font-medium">
          Manage advisory, review, and training consulting projects.
        </p>
      </div>
      <div class="flex items-center gap-3">
        <UButton
          color="primary"
          label="New Assignment"
          icon="i-lucide-plus"
          class="px-6 py-2.5 font-bold rounded-lg shadow-lg"
          @click="store.openForm()"
        />
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-6">
      <UCard class="border-l-4 border-primary-500 shadow-md">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Total Assignments</p>
            <p class="text-2xl font-extrabold text-gray-900 dark:text-white mt-1">{{ store.services.length }}</p>
          </div>
          <div class="p-3 bg-primary-50 dark:bg-primary-950 rounded-full text-primary-500">
            <UIcon name="i-lucide-clipboard-list" class="w-6 h-6" />
          </div>
        </div>
      </UCard>

      <UCard class="border-l-4 border-warning-500 shadow-md">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider">In Progress</p>
            <p class="text-2xl font-extrabold text-gray-900 dark:text-white mt-1">
              {{ store.services.filter(s => s.status === 'In Progress').length }}
            </p>
          </div>
          <div class="p-3 bg-warning-50 dark:bg-warning-950 rounded-full text-warning-500">
            <UIcon name="i-lucide-refresh-cw" class="w-6 h-6 animate-spin-slow" />
          </div>
        </div>
      </UCard>

      <UCard class="border-l-4 border-success-500 shadow-md">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Completed</p>
            <p class="text-2xl font-extrabold text-gray-900 dark:text-white mt-1">
              {{ store.services.filter(s => s.status === 'Completed').length }}
            </p>
          </div>
          <div class="p-3 bg-success-50 dark:bg-success-950 rounded-full text-success-500">
            <UIcon name="i-lucide-check-circle" class="w-6 h-6" />
          </div>
        </div>
      </UCard>

      <UCard class="border-l-4 border-gray-500 shadow-md">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Planned</p>
            <p class="text-2xl font-extrabold text-gray-900 dark:text-white mt-1">
              {{ store.services.filter(s => s.status === 'Planned').length }}
            </p>
          </div>
          <div class="p-3 bg-gray-50 dark:bg-gray-850 rounded-full text-gray-500">
            <UIcon name="i-lucide-calendar" class="w-6 h-6" />
          </div>
        </div>
      </UCard>
    </div>

    <!-- Filters -->
    <UCard class="shadow-sm">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <UFormField label="Search Title">
          <UInput v-model="filters.search" placeholder="Search by project title..." icon="i-lucide-search" class="w-full" />
        </UFormField>
        <UFormField label="Category">
          <USelectMenu v-model="filters.category" :items="['All', ...store.categories]" placeholder="Select Category" class="w-full" />
        </UFormField>
        <UFormField label="Status">
          <USelectMenu v-model="filters.status" :items="['All', ...store.statuses]" placeholder="Select Status" class="w-full" />
        </UFormField>
        <UFormField label="Department">
          <USelectMenu v-model="filters.dept" :items="['All', ...store.departments]" placeholder="Select Department" class="w-full" />
        </UFormField>
      </div>
    </UCard>

    <!-- Data Table -->
    <UCard class="shadow-md overflow-hidden">
      <UTable :data="filteredServices" :columns="columns" class="w-full">
        <!-- Status Cell -->
        <template #status-cell="{ row }">
          <UBadge :color="store.getStatusColor((row.original as any).status)" variant="subtle" size="md">
            {{ (row.original as any).status }}
          </UBadge>
        </template>

        <!-- Attachment Cell -->
        <template #attachment-cell="{ row }">
          <div v-if="(row.original as any).attachment" class="flex items-center gap-1.5">
            <UButton
              icon="i-heroicons-document-arrow-down"
              color="primary"
              variant="ghost"
              size="sm"
              :label="(row.original as any).attachment.name"
              @click="store.downloadAttachment((row.original as any).id, (row.original as any).attachment.name)"
            />
          </div>
          <span v-else class="text-gray-400 italic text-xs">No file</span>
        </template>

        <!-- Actions Cell -->
        <template #actions-cell="{ row }">
          <UButton
            icon="i-lucide-eye"
            color="primary"
            variant="ghost"
            size="sm"
            label="Details"
            @click="store.openDetail(row.original as any)"
          />
        </template>
      </UTable>
    </UCard>

    <!-- Form Modal -->
    <Teleport to="body">
      <div v-if="store.isFormOpen" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">
                {{ store.isEditing ? 'Edit Consulting Assignment' : 'New Consulting Assignment' }}
              </h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.closeForm" />
            </div>
          </template>

          <UForm @submit.prevent="store.saveService" class="space-y-4">
            <UFormField label="Project Title" required>
              <UInput v-model="store.newService.title" placeholder="Advisory on Policy Revisions" class="w-full" required />
            </UFormField>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <UFormField label="Category">
                <USelectMenu v-model="store.newService.category" :items="store.categories" class="w-full" />
              </UFormField>

              <UFormField label="Requestor Department">
                <USelectMenu v-model="store.newService.requestorDept" :items="store.departments" class="w-full" />
              </UFormField>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <UFormField label="Period Quarter">
                <USelectMenu v-model="store.newService.periodQuarter" :items="store.quarters" class="w-full" />
              </UFormField>

              <UFormField label="Period Year">
                <USelectMenu v-model="store.newService.periodYear" :items="store.years" class="w-full" />
              </UFormField>

              <UFormField label="Consultant / Lead Auditor" required>
                <UInput v-model="store.newService.consultantName" placeholder="John Doe" class="w-full" required />
              </UFormField>
            </div>

            <UFormField label="Status">
              <USelectMenu v-model="store.newService.status" :items="store.statuses" class="w-full" />
            </UFormField>

            <UFormField label="Notes / Description">
              <UTextarea v-model="store.newService.notes" placeholder="Detailed notes about the consulting service..." class="w-full" />
            </UFormField>

            <!-- File Upload -->
            <UFormField label="Upload Support Document / Report">
              <div class="border-2 border-dashed border-gray-300 rounded-lg p-6 text-center hover:border-primary-500 transition cursor-pointer relative">
                <input
                  type="file"
                  class="absolute inset-0 opacity-0 cursor-pointer"
                  @change="(e: any) => store.handleFileUpload(e.target.files)"
                />
                <UIcon name="i-lucide-upload-cloud" class="w-10 h-10 text-gray-400 mx-auto mb-2" />
                <p class="text-sm font-medium text-gray-700">
                  {{ store.newService.attachment ? store.newService.attachment.name : 'Click or Drag file to upload' }}
                </p>
                <p class="text-xs text-gray-500 mt-1">PDF or DOCX (max. 10MB)</p>
              </div>
            </UFormField>

            <div v-if="store.errorMsg" class="text-red-500 text-sm font-medium">
              {{ store.errorMsg }}
            </div>

            <div class="flex justify-end gap-3 pt-4 border-t">
              <UButton label="Cancel" color="neutral" variant="ghost" @click="store.closeForm" />
              <UButton type="submit" label="Save Assignment" color="primary" :loading="store.loading" />
            </div>
          </UForm>
        </UCard>
      </div>
    </Teleport>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="store.isDetailOpen" class="fixed inset-0 bg-gray-900/60 z-50 flex items-center justify-center p-4">
        <UCard class="w-full max-w-2xl shadow-2xl">
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-bold text-gray-900 dark:text-white">Assignment Details</h3>
              <UButton icon="i-heroicons-x-mark" color="neutral" variant="ghost" @click="store.closeDetail" />
            </div>
          </template>

          <div v-if="store.selectedService" class="space-y-6">
            <div>
              <h4 class="text-xl font-bold text-gray-900 dark:text-white">{{ store.selectedService.title }}</h4>
              <UBadge :color="store.getStatusColor(store.selectedService.status)" variant="subtle" size="md" class="mt-2">
                {{ store.selectedService.status }}
              </UBadge>
            </div>

            <div class="grid grid-cols-2 gap-4 text-sm bg-gray-50 dark:bg-gray-900 p-4 rounded-lg">
              <div>
                <span class="text-xs text-gray-400 block">Category</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.category }}</span>
              </div>
              <div>
                <span class="text-xs text-gray-400 block">Period</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.period }}</span>
              </div>
              <div>
                <span class="text-xs text-gray-400 block">Requestor Department</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.requestorDept }}</span>
              </div>
              <div>
                <span class="text-xs text-gray-400 block">Consultant Lead</span>
                <span class="font-semibold text-gray-800 dark:text-gray-200">{{ store.selectedService.consultantName }}</span>
              </div>
            </div>

            <div>
              <span class="text-xs text-gray-400 block">Description / Notes</span>
              <p class="text-sm text-gray-700 dark:text-gray-300 mt-1 whitespace-pre-line bg-gray-50 dark:bg-gray-900 p-3 rounded">
                {{ store.selectedService.notes || 'No description provided.' }}
              </p>
            </div>

            <!-- Attachment details -->
            <div v-if="store.selectedService.attachment">
              <span class="text-xs text-gray-400 block mb-2">Attached Document</span>
              <div class="flex items-center justify-between p-3 bg-primary-50 dark:bg-primary-950 rounded-lg border border-primary-100">
                <div class="flex items-center gap-2">
                  <UIcon name="i-lucide-file-text" class="text-primary-500 w-5 h-5" />
                  <div>
                    <span class="font-semibold text-sm text-gray-800 dark:text-gray-200">{{ store.selectedService.attachment.name }}</span>
                    <span class="text-xs text-gray-400 block">{{ store.selectedService.attachment.size }} • Uploaded at {{ store.selectedService.attachment.uploadedAt }}</span>
                  </div>
                </div>
                <UButton
                  icon="i-heroicons-document-arrow-down"
                  label="Download"
                  color="primary"
                  variant="solid"
                  size="sm"
                  @click="store.downloadAttachment(store.selectedService.id, store.selectedService.attachment.name)"
                />
              </div>
            </div>

            <div class="flex justify-between gap-3 pt-4 border-t">
              <UButton
                label="Delete"
                icon="i-lucide-trash"
                color="error"
                variant="ghost"
                @click="confirmDelete"
              />
              <div class="flex gap-2">
                <UButton label="Close" color="neutral" variant="ghost" @click="store.closeDetail" />
                <UButton label="Edit Assignment" color="warning" icon="i-lucide-edit" @click="store.editService" />
              </div>
            </div>
          </div>
        </UCard>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useConsultingServiceStore } from '~/stores/consulting-service'

const store = useConsultingServiceStore()
store.fetchServices()

const columns = [
  { accessorKey: 'title', header: 'Project Title' },
  { accessorKey: 'category', header: 'Category' },
  { accessorKey: 'requestorDept', header: 'Dept' },
  { accessorKey: 'period', header: 'Period' },
  { accessorKey: 'consultantName', header: 'Consultant Lead' },
  { accessorKey: 'status', header: 'Status' },
  { accessorKey: 'attachment', header: 'Document' },
  { accessorKey: 'actions', header: 'Actions' }
]

const filters = reactive({
  search: '',
  category: 'All',
  status: 'All',
  dept: 'All'
})

const filteredServices = computed(() => {
  return store.services.filter(s => {
    const matchesSearch = s.title.toLowerCase().includes(filters.search.toLowerCase()) ||
      s.consultantName.toLowerCase().includes(filters.search.toLowerCase())
    const matchesCategory = filters.category === 'All' || s.category === filters.category
    const matchesStatus = filters.status === 'All' || s.status === filters.status
    const matchesDept = filters.dept === 'All' || s.requestorDept === filters.dept
    return matchesSearch && matchesCategory && matchesStatus && matchesDept
  })
})

const confirmDelete = () => {
  if (confirm('Are you sure you want to delete this consulting assignment?')) {
    store.deleteService()
  }
}
</script>
