## tools

### ideas_log
- description: Log a new idea to the ideas database
- params:
  - name: text
    type: string
    required: true
    description: The idea text to capture

### ideas_list
- description: List all captured ideas
- params:
  - name: limit
    type: number
    required: false
    description: Maximum number of ideas to return (default: 50)

### goals_add
- description: Add a new goal
- params:
  - name: title
    type: string
    required: true
    description: Goal title
  - name: description
    type: string
    required: false
    description: Goal description

### goals_list
- description: List all goals
- params:
  - name: status
    type: string
    required: false
    description: Filter by status (active, completed, paused)

### goals_update
- description: Update a goal's status or description
- params:
  - name: id
    type: string
    required: true
    description: Goal UUID
  - name: status
    type: string
    required: false
    description: New status (active, completed, paused)
  - name: description
    type: string
    required: false
    description: New description

### tasks_add
- description: Add a new task
- params:
  - name: title
    type: string
    required: true
    description: Task title
  - name: goal_id
    type: string
    required: false
    description: UUID of the goal this task belongs to
  - name: due
    type: string
    required: false
    description: Due date in ISO 8601 format (e.g. 2026-03-01)

### tasks_list
- description: List tasks
- params:
  - name: status
    type: string
    required: false
    description: Filter by status (open, done)
  - name: goal_id
    type: string
    required: false
    description: Filter by goal UUID

### tasks_update
- description: Update a task's status or due date
- params:
  - name: id
    type: string
    required: true
    description: Task UUID
  - name: status
    type: string
    required: false
    description: New status (open, done)
  - name: due
    type: string
    required: false
    description: New due date (ISO 8601), or "null" to clear

### schedule_read
- description: Read schedule for a date
- params:
  - name: date
    type: string
    required: false
    description: Date in YYYY-MM-DD format (default: today)

### schedule_write
- description: Write or update schedule for a date
- params:
  - name: date
    type: string
    required: false
    description: Date in YYYY-MM-DD format (default: today)
  - name: focus
    type: string
    required: true
    description: Focus theme for the day
  - name: blocks
    type: string
    required: false
    description: Comma-separated time blocks in format "HH:MM|Activity"
