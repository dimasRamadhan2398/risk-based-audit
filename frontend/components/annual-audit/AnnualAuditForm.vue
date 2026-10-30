<template>
  <UModal
    :open="store.showModal"
    :ui="{
      content: 'w-[calc(100vw-2rem)] sm:w-full sm:max-w-5xl max-h-[92vh] flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-800 rounded-2xl shadow-xl overflow-hidden',
      overlay: 'bg-gray-900/50 dark:bg-black/80 backdrop-blur-md'
    }"
    @update:open="(open: boolean) => { if (!open) store.closeModal() }"
  >
    <template #content>
      <UForm
        :state="store.form"
        class="relative flex max-h-[92vh] flex-col transition-colors duration-300"
        @submit.prevent="store.handleSubmit"
      >
        <!-- Header -->
        <div class="shrink-0 flex items-center justify-between border-b border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
          <div>
            <h3 class="text-lg sm:text-xl font-bold text-[var(--text-main)]">
              {{ store.isEditing ? t('annualAudit.form.updatePlan') : t('annualAudit.form.title') }}
            </h3>
            <p class="mt-0.5 text-xs text-[var(--text-muted)]">
              {{ t('annualAudit.form.title') }} — {{ steps[currentStep]?.label }} ({{ t('auditActivityPlan.form.stepNum', { current: currentStep + 1, total: steps.length }) }})
            </p>
          </div>
          <UButton
            color="neutral"
            variant="ghost"
            icon="i-heroicons-x-mark-20-solid"
            class="-my-1"
            :aria-label="t('common.cancel')"
            @click="store.closeModal"
          />
        </div>

        <!-- Stepper -->
        <div class="w-full overflow-x-auto border-b border-[var(--border-main)] bg-[var(--bg-surface)] px-4 py-3 sm:px-6 sm:py-4 scrollbar-thin">
          <div class="flex min-w-max items-center gap-3 px-1 sm:gap-4">
            <template
              v-for="(step, idx) in steps"
              :key="step.key"
            >
              <div
                class="group flex shrink-0 cursor-pointer select-none items-center gap-2 sm:gap-2.5"
                data-testid="form-step"
                @click="goToStep(idx)"
              >
                <div
                  class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold transition-all duration-300 sm:h-8 sm:w-8"
                  :class="currentStep === idx
                    ? 'bg-primary text-white shadow-md ring-4 ring-primary/20 scale-105'
                    : currentStep > idx
                      ? 'bg-secondary text-white shadow-sm'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-400 group-hover:bg-gray-200 dark:group-hover:bg-gray-700'"
                >
                  <UIcon
                    v-if="currentStep > idx"
                    name="i-heroicons-check"
                    class="h-3.5 w-3.5 stroke-[3]"
                  />
                  <span v-else>{{ idx + 1 }}</span>
                </div>
                <span
                  class="whitespace-nowrap text-xs font-semibold transition-colors"
                  :class="currentStep === idx
                    ? 'text-primary font-bold'
                    : currentStep > idx
                      ? 'text-secondary font-semibold'
                      : 'text-gray-400 dark:text-gray-500 group-hover:text-gray-600 dark:group-hover:text-gray-300'"
                >
                  {{ step.label }}
                </span>
              </div>
              <div
                v-if="idx < steps.length - 1"
                class="h-0.5 w-5 shrink-0 rounded transition-all duration-300 sm:w-8"
                :class="currentStep > idx ? 'bg-secondary' : 'bg-gray-200 dark:bg-gray-800'"
              />
            </template>
          </div>
        </div>

        <!-- Body -->
        <div
          ref="bodyRef"
          class="flex-1 min-h-0 overflow-y-auto p-4 sm:p-6 space-y-6"
        >
          <!-- Step 1: Bagian Plan header -->
          <div
            v-if="currentStep === 0"
            class="space-y-5"
          >
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-document-text"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('annualAudit.form.stepper.header') }}
                  </h3>
                </div>
              </template>

              <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                <UFormField
                  :label="t('annualAudit.form.year')"
                  size="lg"
                  required
                >
                  <USelectMenu
                    v-model="store.form.year"
                    :items="store.yearOptions"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.activityCode')"
                  size="lg"
                  :help="t('annualAudit.form.activityCodeHelp')"
                >
                  <UInput
                    :model-value="activityCodePreview"
                    disabled
                    aria-readonly="true"
                    type="text"
                    :placeholder="t('annualAudit.form.activityCodePending')"
                    class="w-full font-mono text-sm bg-gray-100 dark:bg-gray-800"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.planTitle')"
                  size="lg"
                  class="md:col-span-2"
                >
                  <UInput
                    v-model="store.form.title"
                    type="text"
                    :placeholder="t('annualAudit.form.planTitlePlaceholder')"
                    maxlength="150"
                    class="w-full"
                  />
                </UFormField>
              </div>
            </UCard>
          </div>

          <!-- Step 2: Bagian Activities -->
          <div
            v-else-if="currentStep === 1"
            class="space-y-5"
          >
            <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div class="flex items-center gap-2">
                <UIcon
                  name="i-heroicons-clipboard-document-check"
                  class="h-5 w-5 text-amber-500"
                />
                <h3 class="text-lg font-bold text-[var(--text-main)]">
                  {{ t('annualAudit.form.stepper.activities') }}
                </h3>
              </div>
              <UButton
                :label="t('annualAudit.form.addActivity')"
                icon="i-heroicons-plus"
                color="primary"
                variant="soft"
                class="self-start sm:self-auto font-medium"
                @click="store.addActivity"
              />
            </div>

            <div
              v-if="store.validationErrors.activityDetail"
              class="flex items-start gap-2 rounded-lg border border-error-200 bg-error-50 p-3 text-sm text-error-600"
            >
              <UIcon
                name="i-lucide-circle-alert"
                class="mt-0.5 shrink-0"
              />
              <span>{{ store.validationErrors.activityDetail }}</span>
            </div>

            <div class="space-y-4">
              <UCard
                v-for="(activity, index) in store.form.activities"
                :key="activity.id || index"
                :ui="{ body: 'p-4 sm:p-6' }"
                class="border border-[var(--border-main)] relative"
              >
                <div class="flex justify-between items-center mb-4 border-b border-[var(--border-main)] pb-3">
                  <h5 class="font-bold text-[var(--text-main)]">
                    {{ t('annualAudit.form.subActivity', { number: index + 1 }) }}
                  </h5>
                  <UTooltip
                    v-if="store.form.activities.length > 1"
                    :text="activity.activityPlanRef ? t('annualAudit.form.rowInUseTooltip') : t('common.delete')"
                  >
                    <UButton
                      icon="i-heroicons-trash"
                      color="error"
                      variant="ghost"
                      size="sm"
                      :disabled="!!activity.activityPlanRef"
                      :aria-label="t('common.delete')"
                      @click="store.removeActivity(index)"
                    />
                  </UTooltip>
                </div>

                <!-- The Audit Activity Plan that uses this row -->
                <div
                  v-if="activity.activityPlanRef"
                  class="mb-4 flex items-start gap-2 rounded-lg border border-primary-200 dark:border-primary-800 bg-primary-50 dark:bg-primary-950/40 p-2 text-sm text-primary-700 dark:text-primary-300"
                  data-testid="used-in-activity-plan"
                >
                  <UIcon
                    name="i-heroicons-link"
                    class="mt-0.5 shrink-0"
                  />
                  <span>{{ t('annualAudit.form.usedInActivityPlan', { title: activity.activityPlanRef.activityPlanTitle || '-', code: activity.activityPlanRef.assignmentCode || '-' }) }}</span>
                </div>

                <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
                  <UFormField
                    :label="t('annualAudit.form.objectName')"
                    size="lg"
                  >
                    <UInput
                      v-model="activity.name"
                      type="text"
                      :placeholder="t('annualAudit.form.activityNamePlaceholder')"
                      class="w-full"
                      required
                      maxlength="100"
                      @invalid="($event.target as any)?.setCustomValidity(t('annualAudit.form.activityNameInvalid'))"
                      @input="($event.target as any)?.setCustomValidity('')"
                    />
                    <div class="text-xs text-[var(--text-muted)] mt-1 text-right">
                      {{ activity.name ? activity.name.length : 0 }}/100
                    </div>
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.auditType')"
                    size="lg"
                  >
                    <USelectMenu
                      v-model="(activity.category as any)"
                      :items="categoryOptions"
                      class="w-full"
                    />
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.department')"
                    size="lg"
                  >
                    <USelectMenu
                      v-model="activity.departmentId"
                      :items="store.departments"
                      value-key="id"
                      label-key="name"
                      :loading="store.departmentOptionsLoading"
                      :placeholder="t('annualAudit.form.departmentsPlaceholder')"
                      class="w-full"
                    >
                      <template #empty>
                        {{ store.departmentOptionsLoading ? t('common.loading') : t('annualAudit.form.departmentsEmpty') }}
                      </template>
                    </USelectMenu>
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.auditableEntityLabel')"
                    size="lg"
                    class="md:col-span-2"
                  >
                    <USelectMenu
                      :model-value="auditableEntityValue(activity)"
                      :items="auditableUniverseOptions"
                      label-key="label"
                      by="value"
                      :placeholder="t('annualAudit.form.auditableEntityPlaceholder')"
                      class="w-full"
                      :loading="auditUniverseStore.loading"
                      @update:model-value="(val: any) => handleSelectAuditableEntity(activity, val)"
                    >
                      <template #empty>
                        {{ auditUniverseStore.loading ? t('common.loading') : t('annualAudit.form.auditableEntityEmpty') }}
                      </template>
                      <template #item="{ item }">
                        <div class="flex items-center gap-2 max-w-full w-full py-0.5">
                          <span
                            class="w-2.5 h-2.5 rounded-full shrink-0"
                            :class="getRiskLevelDotClass(item.riskLevel)"
                          />
                          <span class="text-[10px] font-bold text-gray-500 shrink-0">[{{ item.riskLevel }}]</span>
                          <span class="truncate text-sm font-medium">{{ item.label || item.name }}</span>
                          <span
                            v-if="item.riskIndex > 0"
                            class="text-[10px] text-gray-400 shrink-0 ml-1"
                          >({{ (item.riskIndex).toFixed(1) }}%)</span>
                          <UBadge
                            v-if="item.auditPriority"
                            color="error"
                            size="sm"
                            variant="subtle"
                            class="ml-auto"
                          >
                            {{ t('annualAudit.form.priorityBadge') }}
                          </UBadge>
                        </div>
                      </template>
                    </USelectMenu>
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.calculatedRiskLevel')"
                    size="lg"
                    :help="t('annualAudit.form.riskSnapshotHelp')"
                  >
                    <div class="mt-2 flex flex-wrap items-center gap-2">
                      <span class="text-sm font-medium text-gray-700 dark:text-gray-200">
                        {{ activity.riskName || '-' }}
                      </span>
                      <span
                        v-if="activity.riskLevel"
                        class="inline-flex items-center px-2 py-1 rounded-md text-sm font-semibold"
                        :class="getRiskLevelColorClass(activity.riskLevel)"
                      >
                        {{ activity.riskLevel }}
                      </span>
                      <UBadge
                        v-if="auditableUniverseOptions.find(u => u.id === activity.auditUniverseId)?.auditPriority"
                        color="error"
                        size="md"
                        variant="subtle"
                      >
                        {{ t('annualAudit.form.auditPriorityBadge') }}
                      </UBadge>
                    </div>
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.scheduleText')"
                    size="lg"
                  >
                    <UInput
                      v-model="activity.timelineText"
                      type="text"
                      :placeholder="t('annualAudit.form.scheduleTextPlaceholder')"
                      maxlength="150"
                      class="w-full"
                    />
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.teamCount')"
                    size="lg"
                  >
                    <UInput
                      v-model.number="activity.auditorCount"
                      type="number"
                      min="1"
                      max="10"
                      class="w-full"
                    />
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.totalMandays')"
                    size="lg"
                  >
                    <UInput
                      v-model.number="activity.totalMandays"
                      type="number"
                      min="1"
                      class="w-full"
                    />
                  </UFormField>

                  <UFormField
                    :label="t('annualAudit.form.teamLead')"
                    size="lg"
                    class="md:col-span-2"
                  >
                    <USelectMenu
                      v-model="activity.supervisorName"
                      :items="store.supervisorOptions"
                      :loading="store.employeesLoading"
                      :placeholder="t('annualAudit.form.supervisorPlaceholder')"
                      class="w-full"
                      create-item
                      @create="(val: string) => { activity.supervisorName = val.trim() }"
                    >
                      <template #empty>
                        {{ store.employeesLoading ? t('common.loading') : (store.employeesError || t('masterData.employee.empty')) }}
                      </template>
                    </USelectMenu>
                  </UFormField>
                </div>
              </UCard>
            </div>

            <div
              v-if="store.validationErrors.auditor"
              class="flex items-start gap-2 rounded-lg border border-error-200 bg-error-50 p-3 text-sm text-error-600"
            >
              <UIcon
                name="i-lucide-circle-alert"
                class="mt-0.5 shrink-0"
              />
              <span>{{ store.validationErrors.auditor }}</span>
            </div>

            <!-- Totals & Team Capacity Card -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 rounded-xl border border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:p-5">
              <div class="p-3 rounded-lg text-center flex flex-col justify-center border border-primary-500/30 bg-primary-50/50 dark:bg-primary-950/20">
                <span class="text-xs font-semibold uppercase text-[var(--text-muted)]">{{ t('annualAudit.form.totalMandays') }}</span>
                <span class="text-3xl font-extrabold text-primary-600 dark:text-primary-400 my-1">{{ store.totalMandays }}</span>
                <span class="text-xs text-[var(--text-muted)]">{{ t('annualAudit.form.mandaysSumHint') }}</span>
              </div>

              <div class="flex flex-col justify-center p-3">
                <div class="flex justify-between items-center mb-1.5">
                  <label class="text-xs font-semibold uppercase text-[var(--text-muted)]">{{ t('annualAudit.form.teamCapacity') }}</label>
                  <span
                    class="text-xs font-bold"
                    :class="store.utilizationData.color === 'red' ? 'text-error-600' : 'text-success-600'"
                  >
                    {{ store.utilizationData.msg }}
                  </span>
                </div>
                <div class="w-full rounded-full h-2.5 bg-gray-200 dark:bg-gray-700 overflow-hidden">
                  <div
                    class="h-2.5 rounded-full transition-all duration-500"
                    :class="{
                      'bg-success-500': store.utilizationData.color === 'green',
                      'bg-warning-400': store.utilizationData.color === 'yellow',
                      'bg-error-600': store.utilizationData.color === 'red'
                    }"
                    :style="{ width: `${Math.min(store.utilizationData.percent, 100)}%` }"
                  />
                </div>
                <p class="text-xs text-[var(--text-muted)] mt-1.5">
                  {{ t('annualAudit.form.totalLoad', { percent: store.utilizationData.percent.toFixed(1) }) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Step 3: Bagian Timeline dan Notes -->
          <div
            v-else-if="currentStep === 2"
            class="space-y-6"
          >
            <!-- Timeline Card -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-calendar-days"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('annualAudit.form.sectionTimeline') }}
                  </h3>
                </div>
              </template>

              <div
                v-if="store.validationErrors.timeline"
                class="flex items-start gap-2 rounded-lg border border-error-200 bg-error-50 p-3 text-sm text-error-600"
              >
                <UIcon
                  name="i-lucide-circle-alert"
                  class="mt-0.5 shrink-0"
                />
                <span>{{ store.validationErrors.timeline }}</span>
              </div>

              <UFormField
                :label="t('annualAudit.form.selectMonths')"
                size="lg"
                required
              >
                <div
                  class="grid grid-cols-3 md:grid-cols-6 lg:grid-cols-12 gap-2 p-2 sm:p-4"
                  role="group"
                  :aria-label="t('annualAudit.form.selectMonths')"
                  aria-required="true"
                >
                  <div
                    v-for="(month, idx) in store.monthsList"
                    :key="idx"
                    class="cursor-pointer border rounded-lg p-2 text-center text-sm font-semibold transition select-none"
                    :class="store.form.selectedMonths.includes(idx) ? 'border-primary-500 bg-primary-50 dark:bg-primary-950/40 text-primary-700 dark:text-primary-300 shadow-sm font-bold scale-105' : 'border-gray-200 dark:border-gray-800 text-gray-500 hover:border-gray-300 dark:hover:border-gray-700'"
                    @click="store.toggleMonth(idx)"
                  >
                    {{ month }}
                  </div>
                </div>

                <div
                  v-if="store.scheduleWarning"
                  class="flex items-center gap-2 text-warning-600 p-3 mt-3 rounded-lg text-sm border border-warning-200 bg-warning-50 dark:bg-warning-950/20"
                >
                  <UIcon
                    name="i-heroicons-exclamation-triangle"
                    class="h-5 w-5 text-warning-500 shrink-0"
                  />
                  <span>{{ store.scheduleWarning }}</span>
                </div>

                <div class="text-xs text-[var(--text-muted)] p-3 mt-3 rounded-lg border border-[var(--border-main)] bg-[var(--bg-surface)]">
                  <span class="font-bold text-[var(--text-main)]">{{ t('annualAudit.form.quarterDistribution') }}</span>
                  {{ store.computedQuarters.length ? store.computedQuarters.join(', ') : '-' }}
                </div>
              </UFormField>
            </UCard>

            <!-- Notes Card -->
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-4' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-chat-bubble-bottom-center-text"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('annualAudit.form.sectionNotes') }}
                  </h3>
                </div>
              </template>

              <UFormField
                :label="t('annualAudit.form.additionalNote')"
                size="lg"
              >
                <UTextarea
                  v-model="store.form.notes"
                  maxlength="500"
                  :rows="5"
                  :placeholder="t('annualAudit.form.notesPlaceholder')"
                  class="w-full"
                  autoresize
                />
                <div class="flex justify-between mt-1.5 text-xs text-[var(--text-muted)]">
                  <span>{{ t('annualAudit.form.notesHint') }}</span>
                  <span>{{ (store.form.notes || '').length }}/500</span>
                </div>
              </UFormField>
            </UCard>
          </div>

          <!-- Step 4: Bagian Attachment -->
          <div
            v-else-if="currentStep === 3"
            class="space-y-6"
          >
            <UCard
              :ui="{ body: 'p-4 sm:p-6 space-y-5' }"
              class="border border-[var(--border-main)]"
            >
              <template #header>
                <div class="flex items-center gap-2">
                  <UIcon
                    name="i-heroicons-paper-clip"
                    class="h-5 w-5 text-amber-500"
                  />
                  <h3 class="text-lg font-bold text-[var(--text-main)]">
                    {{ t('annualAudit.form.sectionAttachment') }}
                  </h3>
                </div>
              </template>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
                <UFormField
                  :label="t('annualAudit.form.attachmentCategory')"
                  size="lg"
                >
                  <USelectMenu
                    v-model="store.form.attachmentCategory"
                    :items="store.attachmentCategoryOptions"
                    value-key="id"
                    option-key="label"
                    class="w-full"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.attachmentUploadedBy')"
                  size="lg"
                >
                  <UInput
                    v-model="store.form.attachmentUploadedBy"
                    :placeholder="t('annualAudit.form.attachmentUploadedBy')"
                    class="w-full bg-gray-100 dark:bg-gray-800"
                    disabled
                    aria-readonly="true"
                  />
                </UFormField>

                <UFormField
                  :label="t('annualAudit.form.attachmentUploadDate')"
                  size="lg"
                  class="md:col-span-2"
                >
                  <AppDatePicker
                    v-model="store.form.attachmentUploadDate"
                    class="w-full"
                  />
                </UFormField>
              </div>

              <UFormField
                :label="t('annualAudit.form.uploadLabel')"
                size="lg"
              >
                <UFileUpload
                  v-model="store.form.file"
                  layout="list"
                  multiple
                  :label="t('annualAudit.form.dropLabel')"
                  :description="t('annualAudit.form.uploadDescription')"
                  class="w-full"
                  :ui="{
                    base: 'min-h-48'
                  }"
                />
              </UFormField>
            </UCard>
          </div>
        </div>

        <!-- Footer -->
        <div class="shrink-0 flex flex-col-reverse items-stretch justify-between gap-3 border-t border-[var(--border-main)] bg-[var(--bg-surface)] p-4 sm:flex-row sm:items-center">
          <div class="flex flex-col-reverse gap-2 sm:flex-row">
            <UButton
              :label="t('common.cancel')"
              color="neutral"
              variant="ghost"
              class="w-full sm:w-auto"
              @click="store.closeModal"
            />
          </div>

          <div class="flex w-full items-center justify-end gap-2 sm:w-auto sm:gap-3">
            <UButton
              v-if="currentStep > 0"
              icon="i-heroicons-arrow-left"
              :label="t('auditActivityPlan.stepper.previous')"
              color="neutral"
              variant="outline"
              class="flex-1 sm:flex-none"
              @click="currentStep--"
            />
            <UButton
              v-if="currentStep < steps.length - 1"
              trailing-icon="i-heroicons-arrow-right"
              :label="t('auditActivityPlan.stepper.next')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              @click="handleNext"
            />
            <UButton
              v-else
              type="submit"
              icon="i-heroicons-check"
              :label="store.isEditing ? t('annualAudit.form.updatePlan') : t('annualAudit.form.savePlan')"
              color="primary"
              class="flex-1 sm:flex-none font-bold shadow-md shadow-primary/20"
              :disabled="store.utilizationData.color === 'red'"
              @click="handleFormSubmit"
            />
          </div>
        </div>
      </UForm>
    </template>
  </UModal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useI18n } from '~/composables/useI18n'
import { useAnnualPlanStore } from '~/stores/annual-audit'
import { useRiskProfileStore } from '~/stores/risk-profile'
import { useAuditUniverseStore } from '~/stores/audit-universe'
import { useAuthStore } from '~/stores/auth'
import { useToastNotification } from '~/components/shared/ToastNotification.vue'
import { AuditCategory, AuditDepartment, type AuditActivities } from '~/types/audit'
import { getRiskLevelColorClass, getRiskLevelDotClass } from '~/utils/riskLevelBadge'
import {
  buildAuditableEntityOptions,
  findAuditableEntityForActivity,
  NOT_SCORED_RISK_LEVEL,
  resolveAuditableEntitySelection,
  type AuditableEntityOption
} from '~/utils/auditableEntityOptions'

const store = useAnnualPlanStore()
const riskStore = useRiskProfileStore()
const auditUniverseStore = useAuditUniverseStore()
const authStore = useAuthStore()
const toast = useToastNotification()
const { t } = useI18n()

const currentStep = ref(0)
const bodyRef = ref<HTMLElement | null>(null)

const steps = computed(() => [
  { key: 'header', label: t('annualAudit.form.stepper.header') },
  { key: 'activities', label: t('annualAudit.form.stepper.activities') },
  { key: 'timelineNotes', label: t('annualAudit.form.stepper.timelineNotes') },
  { key: 'attachments', label: t('annualAudit.form.stepper.attachments') }
])

watch(() => store.showModal, (isOpen) => {
  if (isOpen) {
    currentStep.value = 0
  }
})

watch(currentStep, () => {
  bodyRef.value?.scrollTo?.({ top: 0, behavior: 'smooth' })
})

const validateStep = (stepIdx: number): boolean => {
  if (stepIdx === 0) {
    if (!store.form.year) {
      toast.showWarning(t('annualAudit.form.year') + ' wajib diisi.')
      return false
    }
  } else if (stepIdx === 1) {
    if (!store.form.activities.length) {
      toast.showWarning('Minimal harus ada 1 aktivitas.')
      return false
    }
    for (let i = 0; i < store.form.activities.length; i++) {
      const act = store.form.activities[i]
      if (!act || !act.name?.trim()) {
        toast.showWarning(t('annualAudit.form.validation.objectRequired', { number: i + 1 }))
        return false
      }
      if (!act.supervisorName?.trim()) {
        toast.showWarning(t('annualAudit.form.validation.leadRequired', { number: i + 1 }))
        return false
      }
    }
  } else if (stepIdx === 2) {
    if (!store.form.selectedMonths || store.form.selectedMonths.length === 0) {
      toast.showWarning(t('annualAudit.form.selectMonths') + ' wajib dipilih.')
      return false
    }
  }
  return true
}

const goToStep = (targetIdx: number) => {
  if (targetIdx <= currentStep.value) {
    currentStep.value = targetIdx
    return
  }
  for (let i = currentStep.value; i < targetIdx; i++) {
    if (!validateStep(i)) return
  }
  currentStep.value = targetIdx
}

const handleNext = () => {
  if (!validateStep(currentStep.value)) return
  currentStep.value++
}

const handleFormSubmit = () => {
  store.handleSubmit()
}

const roleLabels: Record<string, string> = {
  ADMIN: 'System Administrator',
  AUDITOR: 'Auditor',
  DEPARTMENT_HEAD: 'Department Head',
  AUDITEE: 'Auditee',
  VIEWER: 'Viewer',
  EXECUTIVE: 'Chief Audit Executive'
}

const signedInRoleLabel = computed(() => {
  const role = authStore.user?.roles?.[0]
  if (!role) return ''
  return roleLabels[role] || role
    .split('_')
    .filter(Boolean)
    .map(part => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
})

// New plans get their number from the server on save, so the field shows the pattern for the selected year
const activityCodePreview = computed(() => {
  if (store.isEditing) return store.form.code || ''
  return `PKAT-${store.form.year || new Date().getFullYear()}-XXX`
})

const categoryOptions = Object.values(AuditCategory) as AuditCategory[]

riskStore.fetchRisks()

const loadFormOptions = async () => {
  const currentYear = Number(store.form.year) || new Date().getFullYear()
  await Promise.allSettled([
    auditUniverseStore.fetchYearlyUniverse(currentYear),
    auditUniverseStore.fetchCorporateUniverse(),
    store.fetchDepartmentOptions(),
    store.fetchEmployees()
  ])
}

onMounted(() => {
  loadFormOptions()
})

watch([() => store.showModal, signedInRoleLabel], ([isOpen]) => {
  if (isOpen) {
    store.form.attachmentUploadedBy = signedInRoleLabel.value
    loadFormOptions()
  }
})

watch(() => store.form.year, async (newYear) => {
  if (newYear) {
    await auditUniverseStore.fetchYearlyUniverse(Number(newYear))
  }
})

const auditableUniverseOptions = computed<AuditableEntityOption[]>(() =>
  buildAuditableEntityOptions(auditUniverseStore.yearlyUniverse, auditUniverseStore.corporateUniverse)
)

// Pre-selects by auditUniverseId (name only for old rows without an id); the picker shows the row's own
// entity even when it is not in the loaded universe lists.
const auditableEntityValue = (activity: AuditActivities): AuditableEntityOption | undefined =>
  findAuditableEntityForActivity(auditableUniverseOptions.value, activity)

const handleSelectAuditableEntity = (activity: AuditActivities, val: unknown) => {
  // Resolved by corporate id, never by name, so entities sharing a name cannot be confused.
  const match = resolveAuditableEntitySelection(auditableUniverseOptions.value, val)
  if (match) {
    activity.auditUniverseId = match.id || ''
    activity.riskName = match.name
    activity.riskLevel = match.riskLevel || NOT_SCORED_RISK_LEVEL

    // Suggest default activity title if empty or using standard default prefix
    const defaultNamePrefixes = ['Audit Operasional', t('annualAudit.form.defaultActivityName', { entity: '|' }).split('|')[0]!.trim()]
    if (!activity.name || defaultNamePrefixes.some(prefix => activity.name!.startsWith(prefix))) {
      activity.name = t('annualAudit.form.defaultActivityName', { entity: activity.riskName })
    }

    // Auto-detect Department based on entity name
    const nameLower = (activity.riskName || '').toLowerCase()
    let detectedDept = AuditDepartment.OPS
    if (nameLower.includes('it') || nameLower.includes('technology') || nameLower.includes('sistem') || nameLower.includes('cyber') || nameLower.includes('data') || nameLower.includes('security')) {
      detectedDept = AuditDepartment.IT
    } else if (nameLower.includes('finance') || nameLower.includes('keuangan') || nameLower.includes('treasury') || nameLower.includes('tax') || nameLower.includes('pajak') || nameLower.includes('accounting') || nameLower.includes('akuntansi')) {
      detectedDept = AuditDepartment.FINANCE
    } else if (nameLower.includes('hr') || nameLower.includes('human') || nameLower.includes('sumber daya') || nameLower.includes('resource') || nameLower.includes('talent') || nameLower.includes('personalia')) {
      detectedDept = AuditDepartment.HR
    }
    // Only use a department the master data actually returned; the guess is kept only when its name matches one.
    const matchedDept = store.departments.find(d => d.name.toLowerCase() === detectedDept.toLowerCase())
    if (matchedDept) {
      activity.departmentId = matchedDept.id
      activity.department = matchedDept.name
    }
  } else {
    activity.auditUniverseId = ''
    activity.riskName = ''
    activity.riskLevel = ''
  }
}

const _getFilteredRisksForDept = (dept: string) => {
  if (!riskStore.risks || riskStore.risks.length === 0) {
    return []
  }
  return riskStore.risks.filter((r) => {
    if (dept === 'IT') return r.category === 'Technology'
    if (dept === 'Finance') return r.category === 'Financial'
    if (dept === 'HR') return r.category === 'Human Resources'
    if (dept === 'Ops') return ['Operations', 'Compliance', 'Strategic', 'Governance'].includes(r.category)
    return true
  })
}
</script>
