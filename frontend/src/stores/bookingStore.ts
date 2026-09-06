import { create } from "zustand"

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

export const useBookingStore = create<BookingState>((set) => ({
  selectedServiceId: null,
  selectedEmployeeId: null,
  selectedDate: null,
  selectedSlot: null,
  setService: (id) => set({ selectedServiceId: id, selectedEmployeeId: null, selectedDate: null, selectedSlot: null }),
  setDate: (date) => set({ selectedDate: date, selectedEmployeeId: null, selectedSlot: null }),
  selectSlot: (employeeId, slot) => set({ selectedEmployeeId: employeeId, selectedSlot: slot }),
  clearSlot: () => set({ selectedSlot: null }),
  reset: () => set({ selectedServiceId: null, selectedEmployeeId: null, selectedDate: null, selectedSlot: null }),
}))
