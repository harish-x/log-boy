export const parseApiDate = (dateString) => {
  const [year, month, day, hour] = dateString.split("-");
  return new Date(year, month - 1, day, hour || 0);
};


export const getAvailableHours = (date, minDate, maxDate) => {
  const dateStr = date.toISOString().split("T")[0];
  const minDateStr = minDate.toISOString().split("T")[0];
  const maxDateStr = maxDate.toISOString().split("T")[0];

  let startHour = 0;
  let endHour = 23;

  if (dateStr === minDateStr) {
    startHour = minDate.getHours();
  }

  if (dateStr === maxDateStr) {
    endHour = maxDate.getHours();
  }

  return Array.from({ length: endHour - startHour + 1 }, (_, i) => startHour + i);
};