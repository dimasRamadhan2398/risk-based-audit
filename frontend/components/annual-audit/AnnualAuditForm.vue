<template>
    <UModal v-model:open="store.showModal" dismissible class="w-full sm:max-w-4xl">
        <template #content>
        <UForm :state="store.form" @submit.prevent="store.handleSubmit">
            <div class="relative bg-[var(--bg-main)] rounded-xl shadow-2xl flex flex-col max-h-[90vh] border border-[var(--border-main)] transition-colors duration-300">
        
                    <div class="px-6 py-4 border-b border-[var(--border-main)] bg-[var(--bg-surface)] rounded-t-xl flex justify-between items-center transition-colors duration-300">
                        <UIcon name="charter" class="text-primary-500 " size="32"></UIcon>
                        <h3 class="text-lg font-bold text-[var(--text-main)]">Annual Audit Form</h3>
                        <UIcon name="close" @click="store.closeModal" class="text-primary-400 hover:text-primary-600  text-2xl cursor-pointer"></UIcon>
                    </div>

        
                    <div class="p-6 overflow-y-auto space-y-4">  
                        <div class="space-y-4">
                            <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b border-[var(--border-main)] pb-2">1. Activity Detail</h4>
                        
                            <div class="grid grid-cols-1 gap-6 md:grid-cols-2 mb-4">
                                <UFormField label="Status" size="lg">
                                    <USelectMenu 
                                    v-model="store.form.status" 
                                    :items="Object.values(AnnualAuditPlanStatus)" 
                                    placeholder="Select Status"
                                    class="w-full"/>
                                </UFormField>

                                <UFormField label="Activity Code" size="lg">  
                                    <UInput 
                                        v-model="store.form.code"
                                        required 
                                        type="text" 
                                        placeholder="e.g. ASR-01"
                                        class="w-full"
                                    />
                                </UFormField>
                            </div>

                            <div class="space-y-4 mt-6">
                                <UCard 
                                    v-for="(activity, index) in store.form.activities" 
                                    :key="index" 
                                    class="relative bg-gray-50 "
                                >
                                    
                                    <div class="flex justify-between items-center mb-4 border-b border-gray-200  pb-2">
                                        <h5 class="font-bold text-gray-700 ">Sub-Aktivitas {{ index + 1 }}</h5>
                                        <UButton 
                                            v-if="store.form.activities.length > 1"
                                            icon="i-heroicons-trash" 
                                            color="error" 
                                            variant="ghost" 
                                            size="sm"
                                            @click="store.removeActivity(index)" 
                                        />
                                    </div>

                                    <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
                                        <UFormField label="Activity Name" size="lg">
                                            <UInput 
                                                v-model="activity.name" 
                                                type="text" 
                                                placeholder="e.g. Audit Operasional Div. Keuangan"
                                                class="w-full"
                                                required
                                            />
                                        </UFormField>

                                        <UFormField label="Category" size="lg">
                                            <USelectMenu v-model="activity.category" :items="Object.values(AuditCategory)" class="w-full" />
                                        </UFormField>

                                        <UFormField label="Department" size="lg">
                                            <USelectMenu v-model="activity.department" :items="Object.values(AuditDepartment)" class="w-full" />
                                        </UFormField>
                                    </div>
                                </UCard>
                            </div>

                            <div class="flex justify-start pt-2">
                                <UButton 
                                    label="Tambah Aktivitas" 
                                    icon="i-heroicons-plus" 
                                    color="primary" 
                                    variant="soft" 
                                    @click="store.addActivity" 
                                />
                            </div>
                        </div>

                        <div class="space-y-4">
                            <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">2. Timeline</h4>
                            
                            <UFormField label="Year" size="lg">
                                <USelectMenu v-model="store.form.year" :items="store.yearOptions" class="w-full" />
                            </UFormField>
                            <UFormField label="Select Months" size="lg">
                                <div class="grid grid-cols-3 md:grid-cols-6 lg:grid-cols-12 gap-2 p-4">
                                    <div v-for="(month, idx) in store.monthsList" :key="idx" 
                                        @click="store.toggleMonth(idx)"
                                        class="cursor-pointer border rounded-lg p-2 text-center text-xs font-semibold transition select-none"
                                        :class="store.form.selectedMonths.includes(idx) ? 'bg-secondary-600  border-secondary-600 shadow-md transform scale-105' : 'bg-gray-50 text-gray-500 hover:bg-gray-100'"
                                    >
                                    {{ month }}
                                    </div>
                                </div>

                                <div v-if="store.scheduleWarning" class="flex items-center gap-2 text-warning-600 bg-warning-50 p-4 m-4 rounded-lg text-sm border border-warning-200">
                                    <UIcon name="warning" class=" text-warning-500"></UIcon>
                                    {{ store.scheduleWarning }}
                                </div>
                                    <div v-if="store.quarterAlert" class="flex items-center gap-2 text-error-600 bg-red-50 p-4 m-4 rounded-lg text-sm border border-error-200">
                                    <UIcon name="alert" class=" text-error-500"></UIcon>
                                    {{ store.quarterAlert }}
                                </div>
            
                                <div class="text-sm text-gray-600 bg-gray-50 p-4 m-4 rounded border">
                                    <span class="font-bold">Distribusi Triwulan:</span> 
                                        {{ store.computedQuarters.length ? store.computedQuarters.join(', ') : '-' }}
                                </div>
                            </UFormField>
                        </div>

                        <div class="space-y-4">
                            <h4 class="text-sm uppercase tracking-wide text-primary-500 font-bold border-b pb-2">3. Auditor</h4>
            
                            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 p-4 bg-primary-50 rounded-lg border border-primary-200 ">
                            <UFormField label="Number of Auditors (1-10)" size="lg"
                            >
                                <UInput v-model.number="store.form.auditorCount" type="number" min="1" max="10" class="w-full"/>
                                <p class="text-xs text-gray-500 mt-1">Advice: High Risk min. 3 auditor</p>
                            </UFormField>
            
                            <UFormField
                                label="Duration (Days)"
                                size="lg"
                            >
                                <UInput v-model.number="store.form.daysPerAuditor" type="number" min="1" />
                            </UFormField>

                            <UBadge class=" p-3 rounded text-center bg-white flex flex-col justify-center">
                                <span class="text-xs text-gray-500 uppercase">Total Mandays</span>
                                <span class="text-2xl font-bold text-primary-600">{{ store.totalMandays }}</span>
                                <span class="text-xs text-gray-400">= {{ store.form.auditorCount }} person × {{ store.form.daysPerAuditor }} day</span>
                            </UBadge>
                            </div>

                            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                            <UFormField label="Supervisor" size="lg">
              
                                <USelectMenu 
                                    :model-value="(store.form.supervisorId as any)"
                                    @update:model-value="(val: any) => store.form.supervisorId = val"
                                    :items="store.supervisorOptions" 
                                    value-key="id"
                                    option-key="label"
                                    placeholder="-- Choose Supervisor --"
                                    class="w-full rounded-md shadow-sm"
                                />

                                <p v-if="store.selectedSupervisor?.workload! > 6" class="text-xs text-error-500 mt-1 font-bold">
                                    ⚠️ Warning: This Supervisor is supervised > 6 Activity!
                                </p>
                            </UFormField>

                            <UFormField>
                                <div class="flex justify-between items-center mb-1">
                                    <label class="label mb-0">Team Estimation Capacity</label>
                                    <span class="text-xs font-bold" :class="store.utilizationData.color === 'red' ? 'text-error-600' : 'text-success-600'">
                                        {{ store.utilizationData.msg }}
                                    </span>
                                </div>
                                <div class="w-full bg-gray-200 rounded-full h-2.5 ">
                                    <div class="h-2.5 rounded-full transition-all duration-500" 
                                        :class="{
                                        'bg-success-500': store.utilizationData.color === 'green',
                                        'bg-warning-400': store.utilizationData.color === 'yellow',
                                        'bg-error-600': store.utilizationData.color === 'red'
                                        }"
                                        :style="{ width: `${Math.min(store.utilizationData.percent, 100)}%` }"
                                    ></div>
                                </div>
                                <p class="text-xs text-gray-400 mt-1">
                                    Total Load: {{ store.utilizationData.percent.toFixed(1) }}% from Annual Capacity.
                                </p>
                            </UFormField>
                            </div>
                        </div>

                        <div class="space-y-4">
                            <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">4. Notes</h4>
                            <UFormField
                                label="Additional Note (Optional)"
                                size="lg"
                            >
                                <UTextarea 
                                    v-model="store.form.notes" 
                                    maxlength="500"
                                    :rows="5"
                                    placeholder="Example: High Priority - external recommendation..."
                                    class="w-full"
                                    autoresize
                                ></UTextarea>
                                <div class="flex justify-between mt-1">
                                    <span class="text-xs text-gray-400">Use it for Special Audit / Investigation.</span>
                                    <span class="text-xs text-gray-400">{{ store.form.notes!.length }}/500</span>
                                </div>
                            </UFormField>
                        </div>

                        <div class="space-y-4">
                            <h4 class="text-sm uppercase tracking-wide text-gray-500 font-bold border-b pb-2">5. Attachment</h4>
                            <UFormField label="Attachment Category">
                                <USelectMenu v-model="store.form.attachmentCategory" :items="store.attachmentCategoryOptions" value-key="id" option-key="label" class="w-full"/>
                            </UFormField>
                            <UFormField
                                label="Attachment Uploaded By"
                                class="block text-sm font-medium"
                                size="lg"
                            >
                                <UInput
                                    v-model="store.form.attachmentUploadedBy"
                                    placeholder="Attachment Uploaded By"
                                    class="w-full"
                                />
                            </UFormField>
                            <UFormField label="Attachment Upload Date">
                                <UInput type="date" v-model="store.form.attachmentUploadDate" class="w-full"/>
                            </UFormField>
                            <UFormField
                                label="Upload your Attachment here"
                                size="lg"
                            >
                                <UFileUpload
                                    v-model="store.form.file"
                                    layout="list"
                                    multiple
                                    label="Drop your attachments here"
                                    description="You can upload multiple files (max. 2MB each)"
                                    class="w-full"
                                    :ui="{
                                        base: 'min-h-48'
                                    }"
                                />
                            </UFormField>
                        </div>
                    </div>
    
                <div class="px-6 py-4 bg-secondary-50  border-t border-secondary-200  rounded-b-xl flex justify-end gap-3">
                    <UButton 
                        :label="store.isEditing ? 'Update Plan' : 'Save Plan'" 
                        color="primary" 
                        :disabled="!!store.quarterAlert || store.utilizationData.color === 'red'"
                        @click="store.handleSubmit" 
                    />
                </div>
            
            </div>
            </UForm>   
        </template>   
    </UModal>
    
</template>    

<script setup lang="ts">
import { computed } from 'vue'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { AnnualAuditPlanStatus, AuditCategory, AuditDepartment } from '~/types/audit'

// Cukup inisialisasi store. Komponen akan otomatis membaca status showModal, data form, dan fungsi dari sini.
const store = useAnnualPlanStore()

</script>