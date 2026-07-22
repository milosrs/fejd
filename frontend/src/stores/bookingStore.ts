import { create } from "zustand"

export interface BookingState {
  selectedServiceId: string | null
  selectedEmployeeId: string | null
  selectedDate: string | null
  selectedSlot: { start_time: string; end_time: string } | null
  setService: (id: string) => void
  setEmployee: (id: string) => void
  setDate: (date: string) => void
  setSlot: (slot: { start_time: string; end_time: string }) => void
  reset: () => void
}

export const useBookingStore = create<BookingState>((set) => ({
  selectedServiceId: null,
  selectedEmployeeId: null,
  selectedDate: null,
  selectedSlot: null,
  setService: (id) => set({ selectedServiceId: id, selectedEmployeeId: null, selectedDate: null, selectedSlot: null }),
  setEmployee: (id) => set({ selectedEmployeeId: id, selectedDate: null, selectedSlot: null }),
  setDate: (date) => set({ selectedDate: date, selectedSlot: null }),
  setSlot: (slot) => set({ selectedSlot: slot }),
  reset: () => set({ selectedServiceId: null, selectedEmployeeId: null, selectedDate: null, selectedSlot: null }),
}))
