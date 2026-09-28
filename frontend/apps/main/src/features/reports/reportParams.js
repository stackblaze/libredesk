export function reportParams(days, teamId, priorityId) {
  const params = { days }
  if (teamId) params.team_id = teamId
  if (priorityId) params.priority_id = priorityId
  return params
}
