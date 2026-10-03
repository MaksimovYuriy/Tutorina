import { useId, useState } from 'react'
import { Button } from '@mui/material'
import {
  slotStatusLabels,
  type PublicSlot,
  type Slot,
} from '../../services/api'
import { useCurrentTime } from '../../services/time'
import {
  calendarDate,
  clockTime,
  monday,
  moscowDate,
  shiftDate,
} from './calendar'
import './SchoolSchedule.css'
const weekdays = [
  'Понедельник',
  'Вторник',
  'Среда',
  'Четверг',
  'Пятница',
  'Суббота',
  'Воскресенье',
]
const shortDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']
type Props =
  | { slots: PublicSlot[]; onEdit?: never; onDelete?: never }
  | {
      slots: Slot[]
      onEdit: (slot: Slot) => void
      onDelete: (slot: Slot) => void
    }
export function SchoolSchedule(props: Props) {
  const now = useCurrentTime()
  const today = moscowDate(now)
  const [week, setWeek] = useState(() => monday(moscowDate(Date.now())))
  const [selectedDay, setSelectedDay] = useState(
    () => (new Date(`${moscowDate(Date.now())}T12:00:00Z`).getUTCDay() + 6) % 7,
  )
  const headingId = useId()
  const lastDay = shiftDate(week, 6)
  const firstDateLabel = calendarDate(week).replace(/ г\.$/, '')
  const lastDateLabel = calendarDate(lastDay).replace(/ г\.$/, '')
  const weekLabel = `${week.slice(0, 4) === lastDay.slice(0, 4) ? firstDateLabel.replace(/ \d{4}$/, '') : firstDateLabel} — ${lastDateLabel}`
  const dates = weekdays.map((_, index) => shiftDate(week, index))
  const weekSlots = props.slots
    .filter((slot) => dates.includes(moscowDate(Date.parse(slot.startsAt))))
    .sort(
      (a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt) || a.id - b.id,
    )
  const times = weekSlots.length
    ? [...new Set(weekSlots.map((slot) => clockTime(slot.startsAt)))].sort()
    : ['09:00', '11:00', '13:00', '15:00', '17:00', '19:00']
  const nearest = props.slots
    .filter((slot) => Date.parse(slot.startsAt) > now)
    .sort((a, b) => Date.parse(a.startsAt) - Date.parse(b.startsAt))[0]
  function goNearest() {
    if (!nearest) return
    const date = moscowDate(Date.parse(nearest.startsAt))
    setWeek(monday(date))
    setSelectedDay((new Date(`${date}T12:00:00Z`).getUTCDay() + 6) % 7)
  }
  function goToday() {
    setWeek(monday(today))
    setSelectedDay((new Date(`${today}T12:00:00Z`).getUTCDay() + 6) % 7)
  }
  function lesson(slot: PublicSlot | Slot) {
    const admin = 'occupied' in slot
    const tone = admin
      ? slot.status === 'cancelled'
        ? 'cancelled'
        : slot.status === 'completed'
          ? 'completed'
          : slot.occupied === slot.capacity
            ? 'full'
            : !slot.published
              ? 'hidden'
              : 'open'
      : 'open'
    const nextDay =
      moscowDate(Date.parse(slot.startsAt)) !==
      moscowDate(Date.parse(slot.endsAt))
    return (
      <article key={slot.id} className={`school-lesson school-lesson--${tone}`}>
        <div className="school-lesson-time">
          {clockTime(slot.startsAt)}–{clockTime(slot.endsAt)}
          {nextDay && (
            <span>
              {' '}
              · до {calendarDate(moscowDate(Date.parse(slot.endsAt)))}
            </span>
          )}
        </div>
        <h3>{slot.title}</h3>
        {slot.level && (
          <div className="school-lesson-level">Уровень: {slot.level}</div>
        )}
        <div className="school-lesson-details">
          {slot.format === 'online' ? 'Онлайн' : 'Очно'} ·{' '}
          {slot.kind === 'individual' ? 'Индивидуально' : 'Группа'}
        </div>
        <div className="school-lesson-places">
          {admin
            ? `Занято ${slot.occupied} из ${slot.capacity}`
            : `Свободных мест: ${slot.freePlaces}`}
        </div>
        {admin && (
          <div className="school-lesson-status">
            {slotStatusLabels[slot.status]} ·{' '}
            {slot.published &&
            slot.status === 'planned' &&
            slot.occupied < slot.capacity &&
            Date.parse(slot.startsAt) > now
              ? 'На доске'
              : 'Скрыт с доски'}
          </div>
        )}
        {props.onEdit && admin && (
          <div className="school-lesson-actions">
            <Button size="small" onClick={() => props.onEdit(slot)}>
              Изменить
            </Button>
            <Button
              size="small"
              color="error"
              onClick={() => props.onDelete(slot)}
            >
              Удалить
            </Button>
          </div>
        )}
      </article>
    )
  }
  return (
    <section className="school-schedule" aria-labelledby={headingId}>
      <div className="school-schedule-top">
        <div>
          <div className="school-schedule-label">Расписание занятий</div>
          <h2 id={headingId}>{weekLabel}</h2>
          <p>
            Время по Москве · {props.onEdit ? 'Все слоты' : 'Свободные места'}
          </p>
        </div>
        <div className="school-week-controls">
          <Button
            aria-label="Предыдущая неделя"
            onClick={() => setWeek(shiftDate(week, -7))}
          >
            ←
          </Button>
          <Button onClick={goToday}>Сегодня</Button>
          <Button
            aria-label="Следующая неделя"
            onClick={() => setWeek(shiftDate(week, 7))}
          >
            →
          </Button>
        </div>
      </div>
      <div className="school-week-summary">
        <span role="status">
          {weekSlots.length
            ? `Занятий на этой неделе: ${weekSlots.length}`
            : props.onEdit
              ? 'На этой неделе занятий нет.'
              : 'На этой неделе нет свободных слотов.'}
        </span>
        {nearest && (
          <Button size="small" onClick={goNearest}>
            К ближайшему занятию
          </Button>
        )}
      </div>
      <div className="school-desktop">
        <table className="school-table">
          <caption className="school-visually-hidden">
            Расписание на неделю с {calendarDate(week)}. Время по Москве.
          </caption>
          <thead>
            <tr>
              <th scope="col" className="school-time-heading">
                № / Время
              </th>
              {dates.map((date, index) => (
                <th
                  scope="col"
                  key={date}
                  className={date === today ? 'school-today' : ''}
                  aria-current={date === today ? 'date' : undefined}
                >
                  <span>{weekdays[index]}</span>
                  <small>
                    {date.slice(8, 10)}.{date.slice(5, 7)}
                  </small>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {times.map((time, index) => (
              <tr key={time}>
                <th scope="row" className="school-time-heading">
                  <small>{index + 1}</small>
                  <strong>{time}</strong>
                </th>
                {dates.map((date) => {
                  const entries = weekSlots.filter(
                    (slot) =>
                      moscowDate(Date.parse(slot.startsAt)) === date &&
                      clockTime(slot.startsAt) === time,
                  )
                  return (
                    <td
                      key={date}
                      className={date === today ? 'school-today' : ''}
                    >
                      {entries.length ? (
                        <div className="school-cell-lessons">
                          {entries.map(lesson)}
                        </div>
                      ) : (
                        <span
                          className="school-empty-cell"
                          aria-label="Нет занятий"
                        >
                          —
                        </span>
                      )}
                    </td>
                  )
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="school-mobile">
        <div
          className="school-day-picker"
          role="group"
          aria-label="День недели"
        >
          {dates.map((date, index) => (
            <button
              key={date}
              type="button"
              aria-label={`${weekdays[index]}, ${calendarDate(date)}`}
              aria-pressed={index === selectedDay}
              className={index === selectedDay ? 'school-day-selected' : ''}
              onClick={() => setSelectedDay(index)}
            >
              <span>{shortDays[index]}</span>
              <strong>{date.slice(8, 10)}</strong>
              <i aria-hidden="true">{date === today ? '•' : '\u00a0'}</i>
            </button>
          ))}
        </div>
        <h3 className="school-day-heading">
          {weekdays[selectedDay]}, {calendarDate(dates[selectedDay])}
        </h3>
        <div className="school-day-lessons">
          {weekSlots.filter(
            (slot) =>
              moscowDate(Date.parse(slot.startsAt)) === dates[selectedDay],
          ).length ? (
            weekSlots
              .filter(
                (slot) =>
                  moscowDate(Date.parse(slot.startsAt)) === dates[selectedDay],
              )
              .map(lesson)
          ) : (
            <div className="school-day-empty">
              {props.onEdit
                ? 'На этот день занятий нет.'
                : 'На этот день нет свободных слотов.'}
            </div>
          )}
        </div>
      </div>
    </section>
  )
}
