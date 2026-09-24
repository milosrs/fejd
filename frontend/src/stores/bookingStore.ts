import { create } from "zustand"
import { createJSONStorage, persist } from "zustand/middleware"

export interface BookingState {
  selectedServiceId: string | null
  selectedEmployeeId: string | null
  selectedDate: string | null
  selectedSlot: { start_time: string; end_time: string } | null
  setService: (id: string) => void
  setDate: (date: string) => void
  selectSlot: (employeeId: string, slot: { start_time: string; end_time: string }) => void
  clearSlot: () => void
  reset: () => void
}

const inMemoryStorage: Storage = {
  getItem: () => null,
  setItem: () => {},
  removeItem: () => {},
  clear: () => {},
  key: () => null,
  get length() {
    return 0
  },
}

function bookingStorage(): Storage {
  try {
    const storage = window.localStorage
    const probe = "__fejd_booking_probe__"
    storage.setItem(probe, "1")
    storage.removeItem(probe)
    return storage
  } catch {
    return inMemoryStorage
  }
}

export const useBookingStore = create<BookingState>()(
  persist(
    (set) => ({
      selectedServiceId: null,
      selectedEmployeeId: null,
      selectedDate: null,
      selectedSlot: null,
      setService: (id) => set({ selectedServiceId: id, selectedEmployeeId: null, selectedDate: null, selectedSlot: null }),
      setDate: (date) => set({ selectedDate: date, selectedEmployeeId: null, selectedSlot: null }),
      selectSlot: (employeeId, slot) => set({ selectedEmployeeId: employeeId, selectedSlot: slot }),
      clearSlot: () => set({ selectedSlot: null }),
      reset: () => set({ selectedServiceId: null, selectedEmployeeId: null, selectedDate: null, selectedSlot: null }),
    }),
    {
      name: "fejd.booking",
      version: 1,
      storage: createJSONStorage(bookingStorage),
      partialize: (state) => ({
        selectedServiceId: state.selectedServiceId,
        selectedEmployeeId: state.selectedEmployeeId,
        selectedDate: state.selectedDate,
        selectedSlot: state.selectedSlot,
      }),
    },
  ),
)
